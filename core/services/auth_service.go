package services

import (
	"core/repositories"
	"core/utils"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrNoEmailProvided     = errors.New("No email provided")
	ErrNoPasswordProvided  = errors.New("No password provided")
	ErrUserAlreadyExists   = errors.New("User already exists")
	ErrSessionFileNotFound = errors.New("Session file not foundd")
)

const (
	SESSION_FILE_PATH = "./vault_session"
	CURRENT_USER_PATH = "./current_user"
)

type AuthService struct {
	authRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
}

func NewAuthService(client *redis.Client,
	repo repositories.UserRepository,
	sessionRepo repositories.SessionRepository) *AuthService {
	return &AuthService{
		authRepo:    repo,
		sessionRepo: sessionRepo,
	}
}

func (s *AuthService) Register(email, password string) error {
	if email == "" {
		return ErrNoEmailProvided
	}
	if password == "" {
		return ErrNoPasswordProvided
	}

	// Check if user already exists in the database
	_, err := s.authRepo.GetUserPasswordByEmail(context.Background(), email)
	if err == nil {
		// User exists (no error means we found the user)
		return ErrUserAlreadyExists
	}
	// If we get an error other than "user not found", that's a real error
	if !errors.Is(err, sql.ErrNoRows) {
		// Check if the error message contains "user not found" (custom error from repo)
		if !strings.Contains(err.Error(), "user not found") {
			return fmt.Errorf("error checking user existence: %w", err)
		}
	}

	// Hash password (with PBKDF2)
	hash := pbkdf2.Key([]byte(password), []byte("salt"), 1000, 32, sha256.New)
	hashedPassword := hex.EncodeToString(hash)
	userID := utils.GenerateRandomString(16)
	// Store user in the database
	err = s.authRepo.CreateUser(context.Background(), userID, email, hashedPassword)
	if err != nil {
		return err
	}
	fmt.Println("User registered successfully. Try logging in with vault login")
	return nil
}

func (s *AuthService) Login(email, password string) error {
	if email == "" {
		return ErrNoEmailProvided
	}

	if password == "" {
		return ErrNoPasswordProvided
	}

	// Check if user exists in the database
	userHash, err := s.authRepo.GetUserPasswordByEmail(context.Background(), email)
	if err != nil {
		return err
	}
	// Verify password
	hash := pbkdf2.Key([]byte(password), []byte("salt"), 1000, 32, sha256.New)
	hashedInputPassword := hex.EncodeToString(hash)
	if hashedInputPassword != userHash {
		return fmt.Errorf("invalid password")
	}
	// If password matches, create a session
	// Check if session already exists
	sessionExists, err := s.sessionRepo.CheckSessionExists(context.Background(), email)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if sessionExists {
		return fmt.Errorf("session already exists for user %s", email)
	}

	// If session does not exist, create a new session
	// Generate a session ID and store it in Redis
	// Session expires in 15 minutes as required
	sessionID := utils.GenerateRandomString(16)
	exp := 15 * time.Minute
	err = s.sessionRepo.CreateSession(context.Background(), email, sessionID, exp)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	// Store this "state" that the token represents in a file
	file, err := os.Create("./vault_session")
	if err != nil {
		return errors.New("Session file not created")
	}
	defer file.Close()
	_, err = file.WriteString(sessionID)
	if err != nil {
		return errors.New("Something went wrong while writing the token to your session file")
	}
	// Store the current logged in user
	userFile, err := os.Create("./current_user")
	if err != nil {
		return errors.New("Current identity file not created")
	}
	defer userFile.Close()
	_, err = userFile.WriteString(email)
	if err != nil {
		return errors.New("Something went wrong while writing current user's identity to your session file")
	}
	fmt.Printf("Login successful. Your session ID is: %s\n", sessionID)
	return nil
}

// Logout for CLI - deletes session from Redis and removes local files
func (s *AuthService) Logout() error {
	sessionFilePath := "./vault_session"
	currentUserFilePath := "./current_user"
	sessionIDBytes, err := os.ReadFile(sessionFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrSessionFileNotFound
		}
		return fmt.Errorf("failed to read session file %s: %w", sessionFilePath, err)
	}
	sessionID := string(sessionIDBytes)
	deleted, err := s.sessionRepo.DeleteSession(context.Background(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}
	if !deleted {
		// This means the session ID was not found in Redis, perhaps it expired or was already deleted.
		fmt.Printf("Warning: Session ID '%s' not found in Redis, might have expired or already been logged out.\n", sessionID)
	}

	err = os.Remove(sessionFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File might have been manually deleted already, which is fine for logout.
			fmt.Printf("Warning: Session file '%s' already removed locally.\n", sessionFilePath)
		} else {
			return fmt.Errorf("failed to delete local session file %s: %w", sessionFilePath, err)
		}
	}

	err = os.Remove(currentUserFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File might have been manually deleted already, which is fine for logout.
			fmt.Printf("Warning: Current user file '%s' already removed locally.\n", sessionFilePath)
		} else {
			return fmt.Errorf("failed to delete local session file %s: %w", sessionFilePath, err)
		}
	}
	fmt.Println("Logged out successfully.")
	return nil
}

/* API stuff from below here */

// Logs in a user with basic auth
func (s *AuthService) BasicAuthLogin(email, password string) (string, error) {
	if email == "" {
		return "", ErrNoEmailProvided
	}

	if password == "" {
		return "", ErrNoPasswordProvided
	}

	// Check if user exists in the database
	userHash, err := s.authRepo.GetUserPasswordByEmail(context.Background(), email)
	if err != nil {
		log.Printf("Error getting user password for email %s: %v", email, err)
		return "", fmt.Errorf("user not found or database error: %w", err)
	}

	// Verify password
	hash := pbkdf2.Key([]byte(password), []byte("salt"), 1000, 32, sha256.New)
	hashedInputPassword := hex.EncodeToString(hash)
	log.Printf("Password hash comparison - DB: %s, Input: %s", userHash, hashedInputPassword)

	if hashedInputPassword != userHash {
		log.Printf("Password mismatch for user %s", email)
		return "", fmt.Errorf("invalid credentials")
	}

	// If password matches, check existing session and potentially reuse or create new
	sessionExists, err := s.sessionRepo.CheckSessionExists(context.Background(), email)
	if err != nil {
		log.Printf("Error checking session existence: %v", err)
		// Don't fail login just because we can't check session - continue to create new session
	}

	if sessionExists {
		log.Printf("Session already exists for user %s, will create new session anyway", email)
		// For API login, we'll allow creating a new session even if one exists
		// This is different from CLI login which prevents multiple sessions
	}

	// Generate a session ID via base64 auth and store it in Redis
	// Session expires in 15 minutes as required
	payload := email + ":" + password
	sessionID := base64.StdEncoding.EncodeToString([]byte(payload))
	exp := 15 * time.Minute
	err = s.sessionRepo.CreateSession(context.Background(), email, sessionID, exp)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("Login successful for user %s", email)
	return sessionID, nil
}

// Logout for API - deletes session from Redis using provided token
func (s *AuthService) LogoutAPI(sessionToken string) error {
	if sessionToken == "" {
		return errors.New("no session token provided")
	}

	// Delete session from Redis
	deleted, err := s.sessionRepo.DeleteSession(context.Background(), sessionToken)
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	if !deleted {
		// Session token was not found in Redis, might have expired or already been deleted
		return errors.New("session token not found or already expired")
	}

	log.Printf("API logout successful for session token")
	return nil
}


// ValidateToken validates a session token and returns the associated user email
func (s *AuthService) ValidateToken(token string) (string, error) {
	// For base64 encoded session tokens (email:password format)
	// Decode the token to get email:password
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("invalid token format: %w", err)
	}

	// Split email:password
	credentials := string(decoded)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid token structure")
	}

	email := parts[0]
	// Check if session exists for this email
	ctx := context.Background()
	exists, err := s.sessionRepo.CheckSessionExists(ctx, email)
	if err != nil {
		return "", fmt.Errorf("error checking session: %w", err)
	}

	if !exists {
		return "", fmt.Errorf("session not found or expired")
	}

	return email, nil
}
