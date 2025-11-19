package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerID     string
	GRPCPort     string
	HTTPPort     string
	DatabaseURL  string
	JWTSecret    string
	Environment  string
}

func Load() *Config {
	return &Config{
		ServerID:    getEnv("SERVER_ID", "grpc-server-1"),
		GRPCPort:    getEnv("GRPC_PORT", "9001"),
		HTTPPort:    getEnv("HTTP_PORT", "8001"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://authuser:authpass@localhost:5432/authdb?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "meu-secret-super-seguro-compartilhado-123"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Config) Print() {
	fmt.Println("==============================================")
	fmt.Printf("Server ID:    %s\n", c.ServerID)
	fmt.Printf("gRPC Port:    %s\n", c.GRPCPort)
	fmt.Printf("HTTP Port:    %s\n", c.HTTPPort)
	fmt.Printf("Environment:  %s\n", c.Environment)
	fmt.Printf("Database:     Connected\n")
	fmt.Println("==============================================")
}
