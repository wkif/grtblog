package media

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type localStorage struct {
	uploadDir string
}

func newLocalStorage(uploadDir string) *localStorage {
	trimmed := strings.TrimSpace(uploadDir)
	if trimmed == "" {
		trimmed = filepath.Join("storage", "uploads")
	}
	return &localStorage{uploadDir: trimmed}
}

func (s *localStorage) Upload(_ context.Context, storedPath string, file *multipart.FileHeader) error {
	diskPath := s.diskPathFromStored(storedPath)
	if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil {
		return err
	}
	if fileExists(diskPath) {
		return nil
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(diskPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (s *localStorage) Delete(_ context.Context, storedPath string) error {
	return removeFile(s.diskPathFromStored(storedPath))
}

func (s *localStorage) Open(_ context.Context, storedPath string) (io.ReadCloser, error) {
	return os.Open(s.diskPathFromStored(storedPath))
}

func (s *localStorage) Move(fromStoredPath string, toStoredPath string) error {
	fromPath := s.diskPathFromStored(fromStoredPath)
	toPath := s.diskPathFromStored(toStoredPath)
	if fromPath == "" || toPath == "" || fromPath == toPath {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(toPath), 0o755); err != nil {
		return err
	}
	if err := os.Rename(fromPath, toPath); err != nil {
		return err
	}
	return nil
}

func (s *localStorage) PublicURL(storedPath string) string {
	normalized := normalizeStoredPath(storedPath)
	if normalized == "" {
		return ""
	}
	return "/uploads" + normalized
}

func (s *localStorage) diskPathFromStored(storedPath string) string {
	trimmed := strings.TrimSpace(normalizeStoredPath(storedPath))
	if trimmed == "" {
		return ""
	}
	clean := filepath.Clean(trimmed)
	clean = strings.TrimPrefix(clean, string(filepath.Separator))
	uploadDir := filepath.Clean(s.uploadDir)
	if strings.HasPrefix(clean, uploadDir+string(filepath.Separator)) || clean == uploadDir {
		return clean
	}
	return filepath.Join(uploadDir, clean)
}
