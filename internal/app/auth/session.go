package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var (
	// ErrSessionNotFound - session id отсутствует/протух в Redis
	ErrSessionNotFound = errors.New("session not found")
	// ErrInvalidSessionType - тип сессии не соответствует ожидаемому ("access"/"refresh")
	ErrInvalidSessionType = errors.New("invalid session type")
)

// SessionClaims - данные, которые мы храним в Redis по session-id.
// Это аналог claims из JWT, только без подписи/токена — серверная сессия.
type SessionClaims struct {
	UserUUID  string    `json:"user_uuid"`
	Role      string    `json:"role"`
	Type      string    `json:"type"` // "access" или "refresh"
	CreatedAt time.Time `json:"created_at"`
}

// SessionService - генерация/валидация/удаление session-id в Redis.
type SessionService struct {
	client     *redis.Client
	ctx        context.Context
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewSessionService(redis *RedisService, accessTTL, refreshTTL time.Duration) *SessionService {
	return &SessionService{
		client:     redis.client,
		ctx:        redis.ctx,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *SessionService) CreateAccessSession(userUUID, role string) (sessionID string, expiresAt time.Time, err error) {
	return s.createSession("access", userUUID, role, s.accessTTL)
}

func (s *SessionService) CreateRefreshSession(userUUID, role string) (sessionID string, expiresAt time.Time, err error) {
	return s.createSession("refresh", userUUID, role, s.refreshTTL)
}

func (s *SessionService) ValidateAccess(sessionID string) (*SessionClaims, error) {
	return s.validateSession("access", sessionID)
}

func (s *SessionService) ValidateRefresh(sessionID string) (*SessionClaims, error) {
	return s.validateSession("refresh", sessionID)
}

// RotateRefresh - “refresh” для сессий: проверяем refresh-session-id и выдаём новую пару id.
// Старый refresh-session-id удаляем (rotation).
func (s *SessionService) RotateRefresh(refreshSessionID string) (newAccessID, newRefreshID string, accessExpiresAt time.Time, err error) {
	claims, err := s.ValidateRefresh(refreshSessionID)
	if err != nil {
		return "", "", time.Time{}, err
	}

	accessID, accessExp, err := s.CreateAccessSession(claims.UserUUID, claims.Role)
	if err != nil {
		return "", "", time.Time{}, err
	}

	newRefresh, _, err := s.CreateRefreshSession(claims.UserUUID, claims.Role)
	if err != nil {
		return "", "", time.Time{}, err
	}

	// rotation: старый refresh удаляем, access пусть протухает сам
	_ = s.DeleteRefresh(refreshSessionID)

	return accessID, newRefresh, accessExp, nil
}

func (s *SessionService) DeleteAccess(sessionID string) error {
	return s.deleteSession("access", sessionID)
}

func (s *SessionService) DeleteRefresh(sessionID string) error {
	return s.deleteSession("refresh", sessionID)
}

func (s *SessionService) createSession(typ, userUUID, role string, ttl time.Duration) (string, time.Time, error) {
	id := uuid.NewString()
	claims := SessionClaims{
		UserUUID:  userUUID,
		Role:      role,
		Type:      typ,
		CreatedAt: time.Now(),
	}
	b, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	key := s.key(typ, id)
	if err := s.client.Set(s.ctx, key, string(b), ttl).Err(); err != nil {
		return "", time.Time{}, err
	}

	return id, time.Now().Add(ttl), nil
}

func (s *SessionService) validateSession(expectedType, sessionID string) (*SessionClaims, error) {
	key := s.key(expectedType, sessionID)
	val, err := s.client.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	var claims SessionClaims
	if err := json.Unmarshal([]byte(val), &claims); err != nil {
		return nil, err
	}
	if claims.Type != expectedType {
		return nil, ErrInvalidSessionType
	}
	return &claims, nil
}

func (s *SessionService) deleteSession(typ, sessionID string) error {
	key := s.key(typ, sessionID)
	return s.client.Del(s.ctx, key).Err()
}

func (s *SessionService) key(typ, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", typ, sessionID)
}





