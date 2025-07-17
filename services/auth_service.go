package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"filevault/repositories"
	"filevault/utils"
	"fmt"
	"os"
	"time"

	"context"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrNoEmailProvided = errors.New("No email provided")
	ErrNoPasswordProvided = errors.New("No password provided")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrSessionFileNotFound = errors.New("Session file not foundd")
)

type AuthService struct {
	client *redis.Client
	authRepo repositories.UserRepository
}

func NewAuthService(client *redis.Client, repo repositories.UserRepository) *AuthService {
	return &AuthService{
		client: client,
		authRepo: repo,
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
		var hashedPassword string
		err := s.authRepo.GetUserPasswordByEmail(context.Background(), email, &hashedPassword)
		if err != nil {
			return err
		}
		// Verify password
		hash := pbkdf2.Key([]byte(password), []byte("salt"), 1000, 32, sha256.New)
		hashedInputPassword := hex.EncodeToString(hash)
		if hashedInputPassword != hashedPassword {
			return fmt.Errorf("invalid password")
		}
		// If password matches, create a session
		// Check if session already exists
		sessionExists, err := s.client.Exists(context.Background(), email).Result()
		if err != nil {
			return fmt.Errorf("failed to check session existence: %w", err)
		}
		if sessionExists > 0 {
			return fmt.Errorf("session already exists for user %s", email)
		}

		// If session does not exist, create a new session
		// Generate a session ID and store it in Redis
		// Session expires in 15 minutes as required
		sessionID := utils.GenerateRandomString(16)
		err = s.client.Set(context.Background(), sessionID, email, 15 * time.Minute).Err()
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

		fmt.Printf("Login successful. Your session ID is: %s\n", sessionID)
		return nil
}



// Gets the user ID from Redis using the session token
func (s *AuthService) getUserID(sessionToken string, conn *redis.Client) (string, error) {
	ctx := context.Background()
	email, err := conn.Get(ctx, sessionToken).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("Session token does not exist")
		}
		return "", fmt.Errorf("Error getting user ID: %v", err)
	}

	// Query the database to get the user ID
	var id string
	err = s.authRepo.GetUserByEmail(ctx, email, &id)
	if err != nil {
		return "", fmt.Errorf("Error querying user ID: %v", err)
	}
	if id == "" {
		return "", fmt.Errorf("User ID not found for email: %s", email)
	}

	return id, nil
}


func (s *AuthService) Logout() error {
	sessionFilePath := "./vault_session"
	sessionIDBytes, err := os.ReadFile(sessionFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrSessionFileNotFound
		}
		return fmt.Errorf("failed to read session file %s: %w", sessionFilePath, err)
	}
	sessionID := string(sessionIDBytes)
	delCount, err := s.client.Del(context.Background(), sessionID).Result()
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}
	if delCount == 0 {
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
	fmt.Println("Logged out successfully.")
	return nil
}