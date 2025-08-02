package controllers

import (
	"context"
	"core/repositories"
	"core/services"
	worker "core/workers"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// AuthController handles authentication-related endpoints
type AuthController struct {
	authRepo        repositories.UserRepository
	fileRepo        repositories.FileRepository
	sessionRepo     repositories.SessionRepository
	fileService     *services.FileService
	authService     *services.AuthService
	redisClient     *redis.Client
	taskDistributor worker.TaskDistributor
}

// NewAuthController creates a new auth controller with injected dependencies
func NewAuthController(
	authRepo repositories.UserRepository,
	fileRepo repositories.FileRepository,
	sessionRepo repositories.SessionRepository,
	fileService *services.FileService,
	authService *services.AuthService,
	redisClient *redis.Client,
	taskDistributor worker.TaskDistributor,
) *AuthController {
	return &AuthController{
		authRepo:        authRepo,
		fileRepo:        fileRepo,
		sessionRepo:     sessionRepo,
		fileService:     fileService,
		authService:     authService,
		redisClient:     redisClient,
		taskDistributor: taskDistributor,
	}
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration
func (ac *AuthController) Register(c *fiber.Ctx) error {
	// Get the request body
	var req RegisterUserRequest
	log.Print(req)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse body for register request",
		})
	}
	// Validate fields
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email",
		})
	}
	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	// Use the auth service to register the user (this will properly hash the password)
	err := ac.authService.Register(req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User successfully registered",
	})
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles user login
func (ac *AuthController) Login(c *fiber.Ctx) error {
	// Get the request body
	var req LoginUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse body for login request",
		})
	}
	log.Print(req)
	// Validate fields
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email",
		})
	}
	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	// Call the login service
	ctx := context.Background()
	sessionId, err := ac.authService.BasicAuthLogin(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}
	// Create a session
	err = ac.sessionRepo.CreateSession(ctx, req.Email, sessionId, 15 * time.Minute)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user session",
		})
	}

	// Store user info in locals for potential use by middleware/other handlers
	c.Locals("user_email", req.Email)
	c.Locals("session_token", sessionId)
	c.Locals("authenticated", true)

	// Return successful login response
	return c.JSON(fiber.Map{
		"message": "User logged in successfully",
		"token":   sessionId,
	})
}

// Logout handles user logout
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	// Get the Authorization header
	authToken := c.Get("X-Token")
	if authToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	email := c.Locals("user_email").(string)

	// Use the auth service to logout (delete session from Redis)
	err := ac.authService.LogoutAPI(email)
	if err != nil {
		log.Printf("Failed to logout user: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired session token",
		})
	}
	// Delete the state data from c.Locals
	c.Locals("user_email", "")
	c.Locals("session_token", "")
	c.Locals("authenticated", false)

	
	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}

// GetCurrentUser returns current user information
func (ac *AuthController) GetCurrentUser(c *fiber.Ctx) error {
	currentUser := c.Locals("user_email")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"current_user": currentUser,
	})
}
