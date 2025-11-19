package auth

import "time"

// User representa um usuário no sistema
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // nunca expor no JSON
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
}

// LoginRequest é o payload de autenticação
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse é a resposta da autenticação
type LoginResponse struct {
	Token   string `json:"token,omitempty"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ValidateTokenRequest é o payload para validar token
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse é a resposta da validação
type ValidateTokenResponse struct {
	Valid    bool   `json:"valid"`
	UserID   int    `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Message  string `json:"message,omitempty"`
}
