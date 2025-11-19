package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// JWTSecret deve ser compartilhado entre todos os servidores
	// Em produção, carregue de variável de ambiente ou secret manager
	JWTSecret = []byte("meu-secret-super-seguro-compartilhado-123")

	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims representa as informações no JWT
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken cria um novo JWT token para o usuário
func GenerateToken(userID int, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth-grpc-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken valida e parseia um JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return JWTSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken cria um novo token baseado em um token válido (mas possivelmente expirado)
func RefreshToken(oldTokenString string) (string, error) {
	// Parse sem validar expiração
	token, err := jwt.ParseWithClaims(oldTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok {
		// Gerar novo token
		return GenerateToken(claims.UserID, claims.Username)
	}

	return "", ErrInvalidToken
}
