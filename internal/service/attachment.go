package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// AttachmentStorage stores and retrieves attachment content.
type AttachmentStorage interface {
	Store(ctx context.Context, teamID uuid.UUID, filename string, content io.Reader) (path string, err error)
}

// LocalAttachmentStorage stores attachments on the local filesystem.
type LocalAttachmentStorage struct {
	basePath string
}

// NewLocalAttachmentStorage creates a new local attachment storage.
func NewLocalAttachmentStorage(basePath string) *LocalAttachmentStorage {
	return &LocalAttachmentStorage{basePath: basePath}
}

// Store writes the attachment content to a file under basePath/teamID/unique-filename.
func (s *LocalAttachmentStorage) Store(ctx context.Context, teamID uuid.UUID, filename string, content io.Reader) (string, error) {
	dir := filepath.Join(s.basePath, teamID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating attachment directory: %w", err)
	}

	// Unique prefix to avoid collisions.
	safeFilename := uuid.New().String() + "_" + filepath.Base(filename)
	fullPath := filepath.Join(dir, safeFilename)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("creating attachment file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("writing attachment content: %w", err)
	}

	return fullPath, nil
}
