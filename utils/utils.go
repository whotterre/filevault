package utils

import (
	"fmt"
	"math"
	"math/rand"
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