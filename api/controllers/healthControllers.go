package controllers

import (
	"core/repositories"
	"core/services"
	worker "core/workers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// HealthController holds the dependencies for health-related endpoints
type HealthController struct {
	authRepo        repositories.UserRepository
	fileRepo        repositories.FileRepository
	sessionRepo     repositories.SessionRepository
	fileService     *services.FileService
	authService     *services.AuthService
	redisClient     *redis.Client
	taskDistributor worker.TaskDistributor
}

// NewHealthController creates a new health controller with injected dependencies
func NewHealthController(
	authRepo repositories.UserRepository,
	fileRepo repositories.FileRepository,
	sessionRepo repositories.SessionRepository,
	fileService *services.FileService,
	authService *services.AuthService,
	//redisClient *redis.Client,
	taskDistributor worker.TaskDistributor,
) *HealthController {
	return &HealthController{
		authRepo:    authRepo,
		fileRepo:    fileRepo,
		sessionRepo: sessionRepo,
		fileService: fileService,
		authService: authService,
		// redisClient:     redisClient,
		taskDistributor: taskDistributor,
	}
}

// GetStatus checks whether the api is resting (pun intended)
func (hc *HealthController) GetStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "OK"})
}

// Get counts of files and users uploaded
func (hc *HealthController) GetStats(c *fiber.Ctx) error {
	statsMap, err := hc.fileService.GetFileStats()
	if err != nil {
		log.Printf("Failed to get file stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get file stats",
		})
	}
	return c.JSON(fiber.Map{"file_stats": statsMap})
}
