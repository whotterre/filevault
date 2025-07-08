package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"context"

	"github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrNoEmailProvided = errors.New("No email provided")
	ErrNoPasswordProvided = errors.New("No password provided")
	ErrUserAlreadyExists = errors.New("User already exists")
)

type AuthService struct {
	conn *sql.DB
	client *redis.Client
}

func NewAuthService(conn *sql.DB, client *redis.Client) *AuthService {
	return &AuthService{
		conn: conn,
		client: client,
	}
}

func (s *AuthService) Register(email, password string) error {
	if email == "" {
		return ErrNoEmailProvided
	}
	if password == "" {
		return ErrNoPasswordProvided
	}

	// Check if user already exists
	val, err := s.client.Get(context.Background(), email).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if val != "" {
		return ErrUserAlreadyExists
	}

	// Hash password (with PBKDF2)
	println("Hello", password)
	hash := pbkdf2.Key([]byte(password), []byte("salt"), 1000, 32, sha256.New)
	hashedPassword := hex.EncodeToString(hash)
	// Store user in the database 
	query := "INSERT INTO users (email, password) VALUES (?, ?)"
	_, err = s.conn.ExecContext(context.Background(), query, email, hashedPassword)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to register user: %w", err)
	}

	fmt.Println("User registered successfully. Try logging in with vault login")
	return nil
}
