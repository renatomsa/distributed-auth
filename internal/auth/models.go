package auth

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token,omitempty"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type ValidateTokenResponse struct {
	Valid    bool   `json:"valid"`
	UserID   int    `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Message  string `json:"message,omitempty"`
}
