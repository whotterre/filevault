package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type FileRepository interface {
	CreateFile(
		fileId, fileName,
		userId, path,
		fileType string,
		size int64,
		parentId string,
		uploadedAt time.Time) error

	DeleteFile(fileId string) error
	GetUserByEmail(ctx context.Context, email string, idDest *string) error
	GetUserPasswordByEmail(ctx context.Context, email string, hashedPwdDst *string) error
	PublishFile(ctx context.Context, fileId string) error
	UnPublishFile(ctx context.Context, fileId string) error
}

type fileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) FileRepository {
	return &fileRepository{
		db: db,
	}
}

func (r *fileRepository) CreateFile(
	fileId, fileName,
	userId, path,
	fileType string,
	size int64,
	parentId string,
	uploadedAt time.Time) error {

	var parentIdNull sql.NullString
	if parentId != "" {
		parentIdNull = sql.NullString{String: parentId, Valid: true}
	} else {
		parentIdNull = sql.NullString{Valid: false}
	}

	fileRecord, err := r.db.Prepare(`INSERT INTO files (
			id, file_name, user_id, size, file_path, file_type, parent_id, uploaded_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("failed to prepare database statement: %w", err)
	}
	defer fileRecord.Close()
	_, err = fileRecord.Exec(fileId, fileName, userId, size, path, fileType, parentIdNull, uploadedAt)
	if err != nil {
		return fmt.Errorf("failed to execute database statement: %w", err)
	}
	return nil
}

// Deletes all files with the passed fileId
func (r *fileRepository) DeleteFile(fileId string) error {
	query := `DELETE FROM files WHERE id = ?`

	deleteQuery, err := r.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = deleteQuery.Exec(fileId)
	if err != nil {
		return fmt.Errorf("Failed to delete file %w", err)
	}
	return nil
}

func (r *fileRepository) GetUserByEmail(ctx context.Context, email string, idDest *string) error {
	query := "SELECT id FROM users WHERE email = ?"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&idDest)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("No user found with email: %s", email)
		}
		return fmt.Errorf("Error querying user ID: %v", err)
	}
	return nil
}

func (r *fileRepository) GetUserPasswordByEmail(ctx context.Context, email string, hashedPwdDst *string) error {
	query := "SELECT password FROM users WHERE email = ?"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&hashedPwdDst)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to query user: %w", err)
	}
	return nil
}

// Makes a file/folder public
func (r *fileRepository) PublishFile(ctx context.Context, fileId string) error {
	// Check file exists
	// If so, check that value of is_public field
	// If is_public is SQLite true, do nothing
	// Otherwise, set it to true
	var isPublic int
	query := `SELECT is_public FROM files WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, fileId).Scan(&isPublic)
	if err == sql.ErrNoRows {
		return fmt.Errorf("File doesn't exist: %w", err)
	}
	if err != nil {
		return fmt.Errorf("Failed to query file: %w", err)
	}
	if isPublic != 0 {
		// Already public, nothing to do
		return nil
	}
	updateQuery := `UPDATE files SET is_public = 1 WHERE id = ?`
	_, err = r.db.ExecContext(ctx, updateQuery, fileId)
	if err != nil {
		return fmt.Errorf("Failed to update file as public: %w", err)
	}
	return nil
}

// Makes a file/folder private
func (r *fileRepository) UnPublishFile(ctx context.Context, fileId string) error {
	// Check file exists
	// If so, check that value of is_public field
	// If is_public is SQLite false, do nothing
	// Otherwise, set it to false
	var isPublic int
	query := `SELECT is_public FROM files WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, fileId).Scan(&isPublic)
	if err == sql.ErrNoRows {
		return fmt.Errorf("File doesn't exist: %w", err)
	}
	if err != nil {
		return fmt.Errorf("Failed to query file: %w", err)
	}
	if isPublic != 0 {
		return nil
	}
	updateQuery := `UPDATE files SET is_public = 0 WHERE id = ?`
	_, err = r.db.ExecContext(ctx, updateQuery, fileId)
	if err != nil {
		return fmt.Errorf("Failed to update file as public: %w", err)
	}
	return nil
}
