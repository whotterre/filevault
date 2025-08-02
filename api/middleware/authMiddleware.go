package middleware

import (
	"core/services"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authToken := c.Get("X-Token")
		if authToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No authorization header provided"})
		}



		// Validate token and get user info
		userEmail, err := authService.ValidateToken(authToken)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		// Store in locals for use by protected handlers
		c.Locals("user_email", userEmail)
		c.Locals("session_token", authToken)
		c.Locals("authenticated", true)

		return c.Next()
	}
}
