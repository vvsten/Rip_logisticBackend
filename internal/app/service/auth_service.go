package service

import (
	"errors"
	"time"

	"rip-go-app/internal/app/auth"
	"rip-go-app/internal/app/ds"
	"rip-go-app/internal/app/repository"
)

// AuthService - сервис авторизации
type AuthService struct {
	repo           *repository.Repository
	sessionService *auth.SessionService
}

// NewAuthService - создание нового сервиса авторизации
// Лаб8: авторизация по session-id в Redis (без JWT).
func NewAuthService(repo *repository.Repository, sessionService *auth.SessionService) *AuthService {
	return &AuthService{
		repo:           repo,
		sessionService: sessionService,
	}
}

// RegisterRequest - запрос на регистрацию
type RegisterRequest struct {
	Login    string `json:"login" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// LoginRequest - запрос на вход
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse - ответ авторизации
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         ds.User   `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Register - регистрация пользователя
func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Проверяем, что пользователь с таким логином не существует
	_, err := s.repo.GetUserByLogin(req.Login)
	if err == nil {
		return nil, errors.New("user with this login already exists")
	}

	// Устанавливаем роль по умолчанию
	role := req.Role
	if role == "" {
		role = ds.RoleBuyer
	}

	// Создаем пользователя
	user := ds.User{
		Login:    req.Login,
		Email:    req.Email,
		Password: req.Password, // пароль будет захеширован в handler
		Name:     req.Name,
		Phone:    req.Phone,
		Role:     role,
	}

	if err := s.repo.CreateUser(&user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Создаем session-id (access/refresh) в Redis
	accessToken, accessExpiresAt, err := s.sessionService.CreateAccessSession(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to create access session")
	}

	refreshToken, _, err := s.sessionService.CreateRefreshSession(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to create refresh session")
	}

	// Убираем пароль из ответа
	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// Login - вход пользователя
func (s *AuthService) Login(req LoginRequest, hashedPassword string) (*AuthResponse, error) {
	// Получаем пользователя по логину
	user, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Проверяем пароль (хеш уже проверен в handler)
	if user.Password != hashedPassword {
		return nil, errors.New("invalid credentials")
	}

	// Создаем session-id (access/refresh) в Redis
	accessToken, accessExpiresAt, err := s.sessionService.CreateAccessSession(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to create access session")
	}

	refreshToken, _, err := s.sessionService.CreateRefreshSession(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to create refresh session")
	}

	// Убираем пароль из ответа
	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// Logout - выход пользователя
func (s *AuthService) Logout(userUUID, accessToken, refreshToken string) error {
	// Session-id: удаляем access session из Redis.
	// Также можем удалить refresh session, если клиент передал его (в отличие от JWT, это нужно для реального инвалидации).
	_ = userUUID

	var firstErr error
	if accessToken != "" {
		if err := s.sessionService.DeleteAccess(accessToken); err != nil {
			firstErr = err
		}
	}
	if refreshToken != "" {
		if err := s.sessionService.DeleteRefresh(refreshToken); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// RefreshTokens - обновление токенов
func (s *AuthService) RefreshTokens(refreshToken string) (*AuthResponse, error) {
	// Валидируем refresh session-id
	claims, err := s.sessionService.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Получаем пользователя
	user, err := s.repo.GetUserByUUID(claims.UserUUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Rotation: выдаем новые session-id и инвалидируем старый refresh
	accessToken, newRefreshToken, accessExpiresAt, err := s.sessionService.RotateRefresh(refreshToken)
	if err != nil {
		return nil, errors.New("failed to refresh token pair")
	}

	// Убираем пароль из ответа
	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// ValidateAccess - проверка доступа к ресурсу
func (s *AuthService) ValidateAccess(userUUID, resource string) (bool, error) {
	// Stateless JWT: доступ определяется валидностью JWT и ролью в claims.
	// Здесь оставим минимальную проверку “пользователь существует”.
	_, err := s.repo.GetUserByUUID(userUUID)
	if err != nil {
		return false, errors.New("user not found")
	}
	_ = resource
	return true, nil
}

