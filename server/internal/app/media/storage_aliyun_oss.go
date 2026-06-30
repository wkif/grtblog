package media

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"path"
	"strings"

	alioss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/grtsinry43/grtblog-v2/server/internal/config"
)

type aliyunOSSStorage struct {
	client        *alioss.Client
	bucket        string
	prefix        string
	publicBaseURL string
}

func newAliyunOSSStorage(cfg config.OSSConfig) (*aliyunOSSStorage, error) {
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, errors.New("OSS_REGION is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, errors.New("OSS_BUCKET is required")
	}
	if strings.TrimSpace(cfg.PublicBaseURL) == "" {
		return nil, errors.New("OSS_PUBLIC_BASE_URL is required")
	}

	creds := credentials.NewEnvironmentVariableCredentialsProvider()
	if strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.AccessKeySecret) != "" {
		creds = credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(cfg.AccessKeyID),
			strings.TrimSpace(cfg.AccessKeySecret),
			strings.TrimSpace(cfg.SecurityToken),
		)
	}

	ossCfg := alioss.LoadDefaultConfig().
		WithCredentialsProvider(creds).
		WithRegion(strings.TrimSpace(cfg.Region))
	if endpoint := strings.TrimSpace(cfg.Endpoint); endpoint != "" {
		ossCfg = ossCfg.WithEndpoint(endpoint)
	}

	return &aliyunOSSStorage{
		client:        alioss.NewClient(ossCfg),
		bucket:        strings.TrimSpace(cfg.Bucket),
		prefix:        strings.Trim(strings.TrimSpace(cfg.Prefix), "/"),
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
	}, nil
}

func TestAliyunOSSConnection(ctx context.Context, cfg config.OSSConfig) error {
	storage, err := newAliyunOSSStorage(cfg)
	if err != nil {
		return err
	}
	_, err = storage.client.GetBucketStat(ctx, &alioss.GetBucketStatRequest{
		Bucket: alioss.Ptr(storage.bucket),
	})
	return err
}

func (s *aliyunOSSStorage) Upload(ctx context.Context, storedPath string, file *multipart.FileHeader) error {
	if file == nil {
		return errors.New("file is required")
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = mime.TypeByExtension(strings.ToLower(path.Ext(file.Filename)))
	}

	objectKey := s.objectKey(storedPath)
	req := &alioss.PutObjectRequest{
		Bucket:        alioss.Ptr(s.bucket),
		Key:           alioss.Ptr(objectKey),
		Body:          src,
		ContentLength: alioss.Ptr(file.Size),
	}
	if contentType != "" {
		req.ContentType = alioss.Ptr(contentType)
	}
	_, err = s.client.PutObject(ctx, req)
	return err
}

func (s *aliyunOSSStorage) Delete(ctx context.Context, storedPath string) error {
	_, err := s.client.DeleteObject(ctx, &alioss.DeleteObjectRequest{
		Bucket: alioss.Ptr(s.bucket),
		Key:    alioss.Ptr(s.objectKey(storedPath)),
	})
	return err
}

func (s *aliyunOSSStorage) Open(ctx context.Context, storedPath string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &alioss.GetObjectRequest{
		Bucket: alioss.Ptr(s.bucket),
		Key:    alioss.Ptr(s.objectKey(storedPath)),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (s *aliyunOSSStorage) Move(ctx context.Context, fromStoredPath string, toStoredPath string) error {
	fromKey := s.objectKey(fromStoredPath)
	toKey := s.objectKey(toStoredPath)
	if fromKey == "" || toKey == "" || fromKey == toKey {
		return nil
	}
	if _, err := s.client.CopyObject(ctx, &alioss.CopyObjectRequest{
		Bucket:       alioss.Ptr(s.bucket),
		Key:          alioss.Ptr(toKey),
		SourceBucket: alioss.Ptr(s.bucket),
		SourceKey:    alioss.Ptr(fromKey),
	}); err != nil {
		return err
	}
	return s.Delete(ctx, fromStoredPath)
}

func (s *aliyunOSSStorage) PublicURL(storedPath string) string {
	objectKey := s.objectKey(storedPath)
	if objectKey == "" {
		return ""
	}
	return s.publicBaseURL + "/" + objectKey
}

func (s *aliyunOSSStorage) objectKey(storedPath string) string {
	key := objectKeyFromStoredPath(storedPath)
	if s.prefix == "" {
		return key
	}
	return strings.TrimPrefix(path.Join(s.prefix, key), "/")
}

func (s *aliyunOSSStorage) matchPublicURL(publicURL string) (string, bool) {
	base := strings.TrimRight(s.publicBaseURL, "/")
	if base == "" {
		return "", false
	}
	if publicURL != base && !strings.HasPrefix(publicURL, base+"/") {
		return "", false
	}
	key := strings.TrimPrefix(publicURL, base)
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		return "", false
	}
	if s.prefix != "" {
		prefix := strings.Trim(s.prefix, "/") + "/"
		if !strings.HasPrefix(key, prefix) {
			return "", false
		}
		key = strings.TrimPrefix(key, prefix)
	}
	return makeOSSStoredPath(key), true
}
