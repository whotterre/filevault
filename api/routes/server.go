package routes

import (
	"api/controllers"
	"api/middleware"
	"core/repositories"
	"core/services"
	worker "core/workers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"
)

// APIServer holds all the dependencies needed by the API handlers
type APIServer struct {
	authRepo        repositories.UserRepository
	fileRepo        repositories.FileRepository
	sessionRepo     repositories.SessionRepository
	fileService     *services.FileService
	authService     *services.AuthService
	redisClient     *redis.Client
	taskDistributor worker.TaskDistributor
}

// NewAPIServer creates a new API server with all dependencies injected
func NewAPIServer(
	authRepo repositories.UserRepository,
	fileRepo repositories.FileRepository,
	sessionRepo repositories.SessionRepository,
	fileService *services.FileService,
	authService *services.AuthService,
	redisClient *redis.Client,
	taskDistributor worker.TaskDistributor,
) *APIServer {
	return &APIServer{
		authRepo:        authRepo,
		fileRepo:        fileRepo,
		sessionRepo:     sessionRepo,
		fileService:     fileService,
		authService:     authService,
		redisClient:     redisClient,
		taskDistributor: taskDistributor,
	}
}

// InitializeRoutes sets up all the routes with the injected dependencies
func (s *APIServer) InitializeRoutes() *fiber.App {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize controllers with dependencies
	authCtrl := controllers.NewAuthController(s.authRepo, s.fileRepo, s.sessionRepo,
		s.fileService, s.authService, s.redisClient, s.taskDistributor)
	healthCtrl := controllers.NewHealthController(s.authRepo, s.fileRepo, s.sessionRepo,
		s.fileService, s.authService, s.taskDistributor)

	// Public routes (no authentication required)
	public := app.Group("/")
	public.Get("/status", healthCtrl.GetStatus)
	public.Get("/stats", healthCtrl.GetStats)

	// Auth routes
	authRoutes := app.Group("/auth")
	authRoutes.Post("/register", authCtrl.Register)
	authRoutes.Post("/login", authCtrl.Login)
	

	// Protected routes (authentication required)
	protected := app.Group("/api")
	// Add authentication middleware here
	protected.Use(middleware.AuthMiddleware(s.authService))
	protected.Get("/logout", authCtrl.Logout)
	protected.Get("/users/me", authCtrl.GetCurrentUser)
	// 	protected.GET("/files", s.ginHandleFiles)
	// 	protected.POST("/files", s.ginHandleFiles)
	// 	protected.GET("/files/:id", s.ginHandleFileByID)
	// 	protected.DELETE("/files/:id", s.ginHandleFileByID)
	// 	protected.GET("/folders", s.ginHandleFolders)
	// 	protected.POST("/folders", s.ginHandleFolders)
	// 	protected.GET("/thumbnails/:id", s.ginGetThumbnail)
	// }

	return app
}

func (s *APIServer) Start() {
	app := s.InitializeRoutes()
	log.Fatal(app.Listen(":6000"))
}
