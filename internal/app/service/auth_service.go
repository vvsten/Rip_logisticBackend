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
	repo       *repository.Repository
	jwtService *auth.JWTService
}

// NewAuthService - создание нового сервиса авторизации
// Лаб7/требование: авторизация только по JWT, без Redis-сессий.
func NewAuthService(repo *repository.Repository, jwtService *auth.JWTService) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtService: jwtService,
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

	// Генерируем токены
	accessToken, err := s.jwtService.GenerateAccessToken(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Убираем пароль из ответа
	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    time.Now().Add(15 * time.Minute), // время жизни access токена
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

	// Генерируем токены
	accessToken, err := s.jwtService.GenerateAccessToken(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.UUID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Убираем пароль из ответа
	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    time.Now().Add(15 * time.Minute), // время жизни access токена
	}, nil
}

// Logout - выход пользователя
func (s *AuthService) Logout(userUUID, accessToken string) error {
	// Stateless JWT: сервер не хранит сессии. Выход — это “забыть токен” на клиенте.
	// (Можно добавить blacklist по jti, но по требованию убираем Redis.)
	_ = userUUID
	_ = accessToken
	return nil
}

// RefreshTokens - обновление токенов
func (s *AuthService) RefreshTokens(refreshToken string) (*AuthResponse, error) {
	// Валидируем refresh токен
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Получаем пользователя
	user, err := s.repo.GetUserByUUID(claims.UserUUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Генерируем новые токены (stateless, без Redis)
	accessToken, newRefreshToken, err := s.jwtService.RefreshTokenPair(refreshToken)
	if err != nil {
		return nil, errors.New("failed to refresh token pair")
	}

	// Убираем пароль из ответа
	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		ExpiresAt:    time.Now().Add(15 * time.Minute), // время жизни access токена
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

