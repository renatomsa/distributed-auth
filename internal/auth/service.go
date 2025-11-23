package auth

import (
	"errors"
	"fmt"
	"log"
)

type Database interface {
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id int) (*User, error)
	ValidatePassword(username, password string) (*User, error)
}

type Service struct {
	db Database
}

func NewService(db Database) *Service {
	return &Service{db: db}
}

func (s *Service) Authenticate(username, password string) (*LoginResponse, error) {
	log.Printf("[AuthService] Authenticating user: %s", username)

	user, err := s.db.ValidatePassword(username, password)
	if err != nil {
		log.Printf("[AuthService] Authentication failed for %s: %v", username, err)
		return &LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

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

func (s *Service) ValidateToken(tokenString string) (*ValidateTokenResponse, error) {
	log.Printf("[AuthService] Validating token")

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
