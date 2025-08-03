package controllers

import (
	"core/repositories"
	"core/services"
	"core/utils"
	worker "core/workers"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type FileController struct {
	authRepo        repositories.UserRepository
	fileRepo        repositories.FileRepository
	sessionRepo     repositories.SessionRepository
	fileService     *services.FileService
	authService     *services.AuthService
	redisClient     *redis.Client
	taskDistributor worker.TaskDistributor
}

func NewFileController(
	authRepo repositories.UserRepository,
	fileRepo repositories.FileRepository,
	sessionRepo repositories.SessionRepository,
	fileService *services.FileService,
	authService *services.AuthService,
	redisClient *redis.Client,
	taskDistributor worker.TaskDistributor,
) *FileController {
	return &FileController{
		authRepo:        authRepo,
		fileRepo:        fileRepo,
		sessionRepo:     sessionRepo,
		fileService:     fileService,
		authService:     authService,
		redisClient:     redisClient,
		taskDistributor: taskDistributor,
	}
}

type UploadFileRequest struct {
	FileName string `form:"file_name"`
	ParentId string `form:"parent_id"`
}

func (s *FileController) UploadFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse form data",
		})
	}
	// Extract user email from request header
	authToken := c.Get("X-Token")
	email, err := utils.ExtractEmailFromToken(authToken)
	log.Println(email)
	fmt.Println(authToken)
	if err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to get email from token",
		})
	}
	files := form.File["file_data"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	fileNames := form.Value["file_name"]
	parentIDs := form.Value["parent_id"]

	fileName := ""
	if len(fileNames) > 0 {
		fileName = fileNames[0]
	}

	parentID := ""
	if len(parentIDs) > 0 {
		parentID = parentIDs[0]
	}

	file := files[0]
	if file == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File data is missing",
		})
	}
	if s.fileService == nil {
		log.Println("FileService is not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error: FileService not initialized",
		})
	}
	newFileInfo, err := s.fileService.UploadFileAPI(fileName, parentID, email, file)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"file_name":  newFileInfo.FileName,
		"file_type":  newFileInfo.FileType,
		"file_id":    newFileInfo.FileId,
		"parent_id":  newFileInfo.ParentId,
		"created_at": newFileInfo.UploadedAt,
	})
}
