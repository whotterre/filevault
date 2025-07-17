package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrUserAlreadyExists = errors.New("User already exists")
)

type UserRepository interface {
	CreateUser(ctx context.Context, userID, email, hashedPassword string) error 
	GetUserPasswordByEmail(ctx context.Context,email string, hashedPwdDst *string) error 
	GetUserByEmail (ctx context.Context, email string, idDest *string) error 
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db:db}
}

func (r *userRepository) CreateUser(ctx context.Context, userID, email, hashedPassword string) error {
	query := "INSERT INTO users (id, email, password) VALUES (?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, userID, email, hashedPassword)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to register user: %w", err)
	}
	return nil
}

func (r *userRepository) GetUserPasswordByEmail(ctx context.Context,email string, hashedPwdDst *string) error {
	query := "SELECT password FROM users WHERE email = ?"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&hashedPwdDst)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to query user: %w", err)
	}
	return nil 
}

func (r *userRepository) GetUserByEmail (ctx context.Context, email string, idDest *string) error {
	query := "SELECT id FROM users WHERE email = ?"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&idDest)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("No user found with email: %s", email)
		}
		return fmt.Errorf("Error querying user ID: %v", err)
	}
	return nil
}

