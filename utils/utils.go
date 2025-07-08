package utils

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
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
	i := int(math.Floor(math.Log(float64(byteSize))/ math.Log(float64(1024))))
	
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

func GetSessionTokenFromFile() (string,error) {
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
	tokenStr, err := GetSessionTokenFromFile()
	if err != nil {
		fmt.Println("Error getting session token:", err)
		return false
	}
	if tokenStr == "" {
		fmt.Println("Session file is empty.")
		return false
	}

	// Check if it exists in Redis
	ctx := context.Background()
	exists := conn.Exists(ctx, tokenStr).Val()
	if exists == 0 {
		fmt.Println("User token does not exist in Redis.")
		return false
	}
	return true
}

// Gets the user ID from Redis using the session token
func GetUserID(sessionToken string, conn *redis.Client, dbConn *sql.DB) (string, error) {
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
	query := "SELECT id FROM users WHERE email = ?"
	err = dbConn.QueryRowContext(ctx, query, email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("No user found with email: %s", email)
		}
		return "", fmt.Errorf("Error querying user ID: %v", err)
	}
	if id == "" {
		return "", fmt.Errorf("User ID not found for email: %s", email)
	}

	return id, nil
}