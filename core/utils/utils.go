package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

func GetSizeField(byteSize int64) string {
	// Converts the size value to a format like "3.7 MB"
	units := []string{"B", "KB", "MB", "GB", "TB", "EB"}

	if byteSize < 0 {
		return "0 B"
	}
	// log byteSize / log 1024 - gets the number of times we need to divide to get the unit
	i := int(math.Floor(math.Log(float64(byteSize)) / math.Log(float64(1024))))

	if i < 0 {
		i = 0
	} else if i >= len(units) {
		i = len(units) - 1
	}

	fileSize := float64(byteSize) / math.Pow(1024, float64(i))
	if i == 0 {
		return fmt.Sprintf("%d %s", byteSize, units[i])
	}
	return fmt.Sprintf("%.1f %s", fileSize, units[i])
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func GetSessionTokenFromFile() (string, error) {
	// Get user token from file
	sessionFilePath := "./vault_session"

	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		fmt.Println("Session file does not exist.")
		return "", err
	}
	// Read the file content
	token, err := os.ReadFile(sessionFilePath)
	if err != nil {
		fmt.Println("Error reading session file:", err)
		return "", err
	}
	tokenStr := string(token)
	return tokenStr, nil
}

// CheckValidUser checks if the user is valid in Redis
func ValidateUser(conn *redis.Client) bool {
	currentUser, err := GetCurrentUser()
	if err != nil {
		fmt.Println("Error getting session token:", err)
		return false
	}
	if currentUser == "" {
		fmt.Println("Session file is empty.")
		return false
	}

	// Check if it exists in Redis
	ctx := context.Background()
	exists := conn.Exists(ctx, currentUser).Val()
	if exists == 0 {
		fmt.Println("User session has expired or doesn't exist. Please log in again")
		return false
	}
	return true
}

// Gets the current user
func GetCurrentUser() (string, error) {
	const CURRENT_USER_FILE = "./current_user"

	if _, err := os.Stat(CURRENT_USER_FILE); os.IsNotExist(err) {
		fmt.Println("Current user file does not exist.")
		return "", err
	}
	// Read the file content
	data, err := os.ReadFile(CURRENT_USER_FILE)
	if err != nil {
		fmt.Println("Error reading session file:", err)
		return "", err
	}
	currentUser := string(data)
	return currentUser, nil
}

// Write JSON response equiv to Gin's c.JSON()
func WriteJSONResponse(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := fmt.Sprintf(`{"message": "%s"}`, msg)
	w.Write([]byte(response))
}

func ToJSON(data map[string]any) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return []byte(""), errors.New("Failed to serialize data as JSON")
	}
	return jsonData, nil
}