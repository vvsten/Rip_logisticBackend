package ds

import (
	"time"
)

// User - модель пользователя
type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UUID      string    `json:"uuid" gorm:"uniqueIndex;not null"`
	Login     string    `json:"login" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // не возвращаем в JSON
	Name      string    `json:"name" gorm:"not null"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role" gorm:"not null;default:'buyer'"` // buyer, manager, admin
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserRole - роли пользователей
const (
	RoleBuyer   = "buyer"
	RoleManager = "manager"
	RoleAdmin   = "admin"
	// Совместимость (в некоторых местах/БД роль могла называться "moderator")
	RoleModerator = "moderator"
)
