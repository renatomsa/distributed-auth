package auth

import (
	"errors"
	"fmt"
	"log"
)

// Database define a interface que a camada de dados deve implementar
type Database interface {
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id int) (*User, error)
	ValidatePassword(username, password string) (*User, error)
}

// Service implementa a lógica de negócio de autenticação
type Service struct {
	db Database
}

// NewService cria uma nova instância do serviço de autenticação
func NewService(db Database) *Service {
	return &Service{db: db}
}

// Authenticate autentica um usuário e retorna um token JWT
func (s *Service) Authenticate(username, password string) (*LoginResponse, error) {
	log.Printf("[AuthService] Authenticating user: %s", username)

	// Validar entrada
	if username == "" || password == "" {
		log.Printf("[AuthService] Invalid input: username or password empty")
		return &LoginResponse{
			Success: false,
			Message: "Username and password are required",
		}, nil
	}

	// Validar credenciais no banco
	user, err := s.db.ValidatePassword(username, password)
	if err != nil {
		log.Printf("[AuthService] Authentication failed for %s: %v", username, err)
		return &LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Gerar token JWT
	token, err := GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("[AuthService] Failed to generate token: %v", err)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("[AuthService] Authentication successful for user: %s (ID: %d)", username, user.ID)

	return &LoginResponse{
		Token:   token,
		Success: true,
		Message: "Login successful",
	}, nil
}

// ValidateToken valida um token JWT e retorna as informações do usuário
func (s *Service) ValidateToken(tokenString string) (*ValidateTokenResponse, error) {
	log.Printf("[AuthService] Validating token")

	if tokenString == "" {
		return &ValidateTokenResponse{
			Valid:   false,
			Message: "Token is required",
		}, nil
	}

	// Validar token
	claims, err := ValidateToken(tokenString)
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			log.Printf("[AuthService] Token expired")
			return &ValidateTokenResponse{
				Valid:   false,
				Message: "Token expired",
			}, nil
		}

		log.Printf("[AuthService] Invalid token: %v", err)
		return &ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid token",
		}, nil
	}

	// Verificar se usuário ainda existe no banco
	user, err := s.db.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("[AuthService] User not found for token: %v", err)
		return &ValidateTokenResponse{
			Valid:   false,
			Message: "User not found",
		}, nil
	}

	log.Printf("[AuthService] Token valid for user: %s (ID: %d)", user.Username, user.ID)

	return &ValidateTokenResponse{
		Valid:    true,
		UserID:   user.ID,
		Username: user.Username,
		Message:  "Token is valid",
	}, nil
}

// RefreshToken renova um token JWT
func (s *Service) RefreshToken(oldToken string) (*LoginResponse, error) {
	log.Printf("[AuthService] Refreshing token")

	if oldToken == "" {
		return &LoginResponse{
			Success: false,
			Message: "Token is required",
		}, nil
	}

	// Gerar novo token
	newToken, err := RefreshToken(oldToken)
	if err != nil {
		log.Printf("[AuthService] Failed to refresh token: %v", err)
		return &LoginResponse{
			Success: false,
			Message: "Failed to refresh token",
		}, nil
	}

	log.Printf("[AuthService] Token refreshed successfully")

	return &LoginResponse{
		Token:   newToken,
		Success: true,
		Message: "Token refreshed",
	}, nil
}
