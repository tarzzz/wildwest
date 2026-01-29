package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role represents user role enum
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser, RoleGuest:
		return true
	default:
		return false
	}
}

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}

// User represents the user domain entity
type User struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Email            string     `json:"email" db:"email" binding:"required,email,max=255"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	Name             string     `json:"name" db:"name" binding:"required,min=2,max=255"`
	Bio              *string    `json:"bio,omitempty" db:"bio" binding:"omitempty,max=5000"`
	AvatarURL        *string    `json:"avatar_url,omitempty" db:"avatar_url" binding:"omitempty,url,max=500"`
	Role             Role       `json:"role" db:"role"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	LastLogin        *time.Time `json:"last_login,omitempty" db:"last_login"`
	FailedLoginCount int        `json:"-" db:"failed_login_count"`
	LockedUntil      *time.Time `json:"-" db:"locked_until"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsGuest checks if the user has guest role
func (u *User) IsGuest() bool {
	return u.Role == RoleGuest
}

// IsDeleted checks if the user has been soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// IsLocked checks if the user account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CanLogin checks if the user can login (active, not locked, not deleted)
func (u *User) CanLogin() bool {
	return u.IsActive && !u.IsLocked() && !u.IsDeleted()
}

// HasRole checks if the user has the specified role
func (u *User) HasRole(role Role) bool {
	return u.Role == role
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name" binding:"required,min=2,max=255"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserRequest represents the user update request
type UpdateUserRequest struct {
	Name      *string `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Bio       *string `json:"bio,omitempty" binding:"omitempty,max=5000"`
	AvatarURL *string `json:"avatar_url,omitempty" binding:"omitempty,url,max=500"`
}

// UpdateUserAdminRequest represents the admin user update request (includes restricted fields)
type UpdateUserAdminRequest struct {
	Name          *string `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Bio           *string `json:"bio,omitempty" binding:"omitempty,max=5000"`
	AvatarURL     *string `json:"avatar_url,omitempty" binding:"omitempty,url,max=500"`
	Role          *Role   `json:"role,omitempty" binding:"omitempty,oneof=admin user guest"`
	IsActive      *bool   `json:"is_active,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
}

// ChangePasswordRequest represents the password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// VerifyEmailRequest represents the email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// UserResponse represents the user in API responses (excludes sensitive fields)
type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Bio             *string    `json:"bio,omitempty"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	Role            Role       `json:"role"`
	IsActive        bool       `json:"is_active"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLogin       *time.Time `json:"last_login,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		Name:            u.Name,
		Bio:             u.Bio,
		AvatarURL:       u.AvatarURL,
		Role:            u.Role,
		IsActive:        u.IsActive,
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		LastLogin:       u.LastLogin,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// ListUsersResponse represents the list users response with pagination
type ListUsersResponse struct {
	Data       []*UserResponse    `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// TokenResponse represents JWT tokens in API responses
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// AuthResponse represents authentication response (login/register)
type AuthResponse struct {
	User   *UserResponse `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

// ErrorResponse represents an error in API responses
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message, requestID string, details map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			Details:   details,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestID,
		},
	}
}
