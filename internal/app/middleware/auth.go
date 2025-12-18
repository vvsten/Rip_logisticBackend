package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"rip-go-app/internal/app/auth"
	"rip-go-app/internal/app/ds"
)

// AuthMiddleware - middleware для проверки авторизации
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware - создание нового middleware
// Лаб7/требование: авторизация только по JWT, без Redis-сессий/blacklist.
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth - middleware для обязательной авторизации
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Валидируем токен
		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		if claims.Type != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid token type",
			})
			c.Abort()
			return
		}

		// Сохраняем информацию о пользователе в контексте
		c.Set("user_uuid", claims.UserUUID)
		c.Set("user_role", claims.Role)
		c.Set("token_claims", claims)

		c.Next()
	}
}

// RequireRole - middleware для проверки роли
func (am *AuthMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Invalid role type",
			})
			c.Abort()
			return
		}

		// Проверяем, есть ли роль пользователя в списке разрешенных
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Insufficient permissions",
		})
		c.Abort()
	}
}

// RequireModerator - middleware для модераторов (Manager или Admin)
func (am *AuthMiddleware) RequireModerator() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Сначала проверяем авторизацию
		token := am.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Валидируем токен
		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		if claims.Type != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid token type",
			})
			c.Abort()
			return
		}

		// Проверяем роль
		if claims.Role != ds.RoleManager && claims.Role != ds.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "Insufficient permissions - only moderators can complete orders",
			})
			c.Abort()
			return
		}

		// Сохраняем информацию о пользователе в контексте
		c.Set("user_uuid", claims.UserUUID)
		c.Set("user_role", claims.Role)
		c.Set("token_claims", claims)

		c.Next()
	})
}

// OptionalAuth - middleware для опциональной авторизации
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		// Валидируем токен
		claims, err := am.jwtService.ValidateToken(token)
		if err != nil || claims.Type != "access" {
			c.Next()
			return
		}

		// Сохраняем информацию о пользователе в контексте
		c.Set("user_uuid", claims.UserUUID)
		c.Set("user_role", claims.Role)
		c.Set("token_claims", claims)

		c.Next()
	}
}

// extractToken - извлечение токена из заголовка Authorization
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Проверяем формат "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetUserUUID - получение UUID пользователя из контекста
func GetUserUUID(c *gin.Context) (string, bool) {
	userUUID, exists := c.Get("user_uuid")
	if !exists {
		return "", false
	}

	uuid, ok := userUUID.(string)
	return uuid, ok
}

// GetUserRole - получение роли пользователя из контекста
func GetUserRole(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := userRole.(string)
	return role, ok
}

// IsAuthenticated - проверка, авторизован ли пользователь
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_uuid")
	return exists
}
