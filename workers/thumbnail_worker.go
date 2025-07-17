package worker

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/prplecake/go-thumbnail"
)

const THUMBNAIL_DIR_PATH = "./storage/thumbnails"
const TaskThumbnailGeneration = "thumbnail:generate"

func GenerateThumbnail(imagePath, fileID string) error {
    config := thumbnail.Generator{
        Width:  100,
        Height: 100,
    }
    tGen := thumbnail.NewGenerator(config)
    image, err := tGen.NewImageFromFile(imagePath)
    if err != nil {
        return err
    }
    
    tNailData, err := tGen.CreateThumbnail(image)
    if err != nil {
        return err
    }

    err = SaveThumbnailToFile(tNailData, fileID)
    if err != nil {
        return err
    }

    return nil
}

func SaveThumbnailToFile(thumbnailData []byte, fileID string) error {
    // Create thumbnails directory if it doesn't exist
    err := os.MkdirAll(THUMBNAIL_DIR_PATH, os.ModePerm)
    if err != nil {
        return fmt.Errorf("failed to create thumbnail directory: %w", err)
    }

    // Save with unique filename based on fileID
    filePath := filepath.Join(THUMBNAIL_DIR_PATH, fmt.Sprintf("%s_thumbnail.png", fileID))
    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create thumbnail file: %w", err)
    }
    defer file.Close()

    _, err = file.Write(thumbnailData)
    if err != nil {
        return fmt.Errorf("failed to write thumbnail data: %w", err)
    }

    return nil
}