package media

import (
	"context"
	"io"
	"mime/multipart"
	"path"
	"strings"
)

const (
	storageProviderLocal     = "local"
	storageProviderAliyunOSS = "aliyun_oss"
	ossStoredPathPrefix      = "oss:"
)

type storageBackend interface {
	Upload(ctx context.Context, storedPath string, file *multipart.FileHeader) error
	Delete(ctx context.Context, storedPath string) error
	Open(ctx context.Context, storedPath string) (io.ReadCloser, error)
	PublicURL(storedPath string) string
}

func normalizeStoredPath(storedPath string) string {
	trimmed := strings.TrimSpace(storedPath)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, ossStoredPathPrefix) {
		key := strings.TrimPrefix(trimmed, ossStoredPathPrefix)
		key = strings.TrimPrefix(path.Clean("/"+key), "/")
		if key == "." {
			return ""
		}
		return ossStoredPathPrefix + key
	}
	clean := path.Clean("/" + strings.TrimPrefix(trimmed, "/"))
	if clean == "/" {
		return ""
	}
	return clean
}

func isOSSStoredPath(storedPath string) bool {
	return strings.HasPrefix(normalizeStoredPath(storedPath), ossStoredPathPrefix)
}

func makeOSSStoredPath(objectKey string) string {
	key := strings.TrimPrefix(path.Clean("/"+strings.TrimSpace(objectKey)), "/")
	if key == "." || key == "" {
		return ""
	}
	return ossStoredPathPrefix + key
}

func objectKeyFromStoredPath(storedPath string) string {
	normalized := normalizeStoredPath(storedPath)
	if !strings.HasPrefix(normalized, ossStoredPathPrefix) {
		return ""
	}
	return strings.TrimPrefix(normalized, ossStoredPathPrefix)
}
