package middleware

import (
	"core/services"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No authorization header provided"})
		}

		var token string

		if !strings.HasPrefix(authHeader, "Basic ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization type"})
		}
		token = strings.TrimPrefix(authHeader, "Basic ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization format"})
		}

		// Validate token and get user info
		userEmail, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		// Store in locals for use by protected handlers
		c.Locals("user_email", userEmail)
		c.Locals("session_token", token)
		c.Locals("authenticated", true)

		return c.Next()
	}
}
