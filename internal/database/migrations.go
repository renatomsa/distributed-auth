package database

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_username ON users(username);
`

func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func SeedUsers(db *sql.DB) error {
	log.Println("Seeding database with test users...")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count > 0 {
		log.Printf("Database already has %d users, skipping seed", count)
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users := []struct {
		username string
		email    string
	}{
		{"alice", "alice@example.com"},
		{"bob", "bob@example.com"},
		{"charlie", "charlie@example.com"},
		{"admin", "admin@example.com"},
	}

	for _, user := range users {
		_, err := db.Exec(
			"INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)",
			user.username, hashedPassword, user.email,
		)
		if err != nil {
			return fmt.Errorf("failed to seed user %s: %w", user.username, err)
		}
		log.Printf("Created user: %s", user.username)
	}

	log.Println("Database seeding completed successfully")
	log.Println("Test credentials: username=alice, password=password123")
	return nil
}
