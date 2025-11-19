package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/renatomsa/auth-grpc/internal/auth"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(connString string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) GetDB() *sql.DB {
	return p.db
}

func (p *PostgresDB) GetUserByUsername(username string) (*auth.User, error) {
	query := `
		SELECT id, username, password_hash, email, created_at
		FROM users
		WHERE username = $1
	`

	var user auth.User
	err := p.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (p *PostgresDB) GetUserByID(id int) (*auth.User, error) {
	query := `
		SELECT id, username, password_hash, email, created_at
		FROM users
		WHERE id = $1
	`

	var user auth.User
	err := p.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (p *PostgresDB) CreateUser(username, password, email string) (*auth.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (username, password_hash, email)
		VALUES ($1, $2, $3)
		RETURNING id, username, password_hash, email, created_at
	`

	var user auth.User
	err = p.db.QueryRow(query, username, hashedPassword, email).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (p *PostgresDB) ValidatePassword(username, password string) (*auth.User, error) {
	user, err := p.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}
