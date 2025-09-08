package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	TOTPSecret   *string    `json:"-" db:"totp_secret"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	Roles        []Role     `json:"roles,omitempty"`
}

type Role struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type Permission struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Resource string    `json:"resource" db:"resource"`
	Action   string    `json:"action" db:"action"`
}

type Secret struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	EncryptedData string    `json:"-" db:"encrypted_data"`
	CreatedBy     uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type AuditLog struct {
	ID         uuid.UUID   `json:"id" db:"id"`
	UserID     *uuid.UUID  `json:"user_id" db:"user_id"`
	Action     string      `json:"action" db:"action"`
	Resource   string      `json:"resource" db:"resource"`
	ResourceID *uuid.UUID  `json:"resource_id" db:"resource_id"`
	Details    interface{} `json:"details" db:"details"`
	IPAddress  string      `json:"ip_address" db:"ip_address"`
	UserAgent  string      `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateSecretRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Data        string `json:"data"`
}