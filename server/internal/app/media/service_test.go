package media

import (
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	domainmedia "github.com/grtsinry43/grtblog-v2/server/internal/domain/media"
)

func TestSyncIndexCreatesAndDeletesRecords(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()
	writePNG(t, filepath.Join(uploadDir, "pictures", "sample.png"))
	writeText(t, filepath.Join(uploadDir, "files", "readme.txt"), "hello")
	writeText(t, filepath.Join(uploadDir, "thumbnails", "pictures", "ignored.txt"), "thumb")

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   99,
		Name: "stale.txt",
		Path: "/files/stale.txt",
		Type: "file",
		Size: 1,
		Hash: "stale-hash",
	})

	svc := NewService(repo, config.StorageConfig{UploadDir: uploadDir}, nil, nil)
	result, err := svc.SyncIndex(context.Background())
	if err != nil {
		t.Fatalf("SyncIndex() error = %v", err)
	}

	if result.Scanned != 2 {
		t.Fatalf("SyncIndex() scanned = %d, want 2", result.Scanned)
	}
	if result.Indexed != 2 {
		t.Fatalf("SyncIndex() indexed = %d, want 2", result.Indexed)
	}
	if result.Created != 2 {
		t.Fatalf("SyncIndex() created = %d, want 2", result.Created)
	}
	if result.Deleted != 1 {
		t.Fatalf("SyncIndex() deleted = %d, want 1", result.Deleted)
	}

	files, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("ListAll() len = %d, want 2", len(files))
	}

	byPath := make(map[string]domainmedia.UploadFile, len(files))
	for _, file := range files {
		byPath[file.Path] = file
	}

	picture, ok := byPath["/pictures/sample.png"]
	if !ok {
		t.Fatalf("missing synced picture record")
	}
	if picture.Type != "picture" {
		t.Fatalf("picture.Type = %q, want picture", picture.Type)
	}

	text, ok := byPath["/files/readme.txt"]
	if !ok {
		t.Fatalf("missing synced file record")
	}
	if text.Type != "file" {
		t.Fatalf("text.Type = %q, want file", text.Type)
	}
}

func TestSyncIndexUpdatesExistingRecordMetadata(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()
	imagePath := filepath.Join(uploadDir, "files", "legacy-photo.jpg")
	writePNG(t, imagePath)

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   7,
		Name: "legacy.bin",
		Path: "/files/legacy-photo.jpg",
		Type: "file",
		Size: 1,
		Hash: "",
	})

	svc := NewService(repo, config.StorageConfig{UploadDir: uploadDir}, nil, nil)
	result, err := svc.SyncIndex(context.Background())
	if err != nil {
		t.Fatalf("SyncIndex() error = %v", err)
	}

	if result.Created != 0 {
		t.Fatalf("SyncIndex() created = %d, want 0", result.Created)
	}
	if result.Updated != 1 {
		t.Fatalf("SyncIndex() updated = %d, want 1", result.Updated)
	}

	files, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("ListAll() len = %d, want 1", len(files))
	}

	file := files[0]
	if file.Name != "legacy-photo.jpg" {
		t.Fatalf("file.Name = %q, want legacy-photo.jpg", file.Name)
	}
	if file.Type != "picture" {
		t.Fatalf("file.Type = %q, want picture", file.Type)
	}
	if file.Hash == "" {
		t.Fatalf("file.Hash should be populated")
	}
	if file.Size <= 1 {
		t.Fatalf("file.Size = %d, want > 1", file.Size)
	}
}

func TestRewriteDraftURLsPromotesCacheAsset(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()
	cacheDiskPath := filepath.Join(uploadDir, "blog", "cache", "draft-image.png")
	writePNG(t, cacheDiskPath)

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   1,
		Name: "draft-image.png",
		Path: "/blog/cache/draft-image.png",
		Type: "picture",
		Size: 128,
		Hash: "draft-hash",
	})

	svc := NewService(repo, config.StorageConfig{UploadDir: uploadDir}, nil, nil)
	rewritten, err := svc.RewriteDraftURLs(context.Background(), "![draft](/uploads/blog/cache/draft-image.png)")
	if err != nil {
		t.Fatalf("RewriteDraftURLs() error = %v", err)
	}
	wantURL := "/uploads/blog/images/draft-image.png"
	if rewritten != "![draft]("+wantURL+")" {
		t.Fatalf("RewriteDraftURLs() = %q, want %q", rewritten, "![draft]("+wantURL+")")
	}
	if _, err := os.Stat(cacheDiskPath); !os.IsNotExist(err) {
		t.Fatalf("expected cache file to be moved, stat err = %v", err)
	}
	finalDiskPath := filepath.Join(uploadDir, "blog", "images", "draft-image.png")
	if _, err := os.Stat(finalDiskPath); err != nil {
		t.Fatalf("expected promoted file at %q: %v", finalDiskPath, err)
	}
	file, err := repo.FindByPath(context.Background(), "/blog/images/draft-image.png")
	if err != nil {
		t.Fatalf("FindByPath(promoted) error = %v", err)
	}
	if file.Type != "picture" {
		t.Fatalf("promoted file.Type = %q, want picture", file.Type)
	}
}

func TestStoredPathFromPublicURLNormalizesWrappedOSSURL(t *testing.T) {
	t.Parallel()

	sysCfg := sysconfig.NewService(&fakeSysConfigStore{
		items: map[string]domainconfig.SysConfig{
			"storage.provider": {
				Key:   "storage.provider",
				Value: "aliyun_oss",
			},
			"storage.oss.region": {
				Key:   "storage.oss.region",
				Value: "cn-beijing",
			},
			"storage.oss.bucket": {
				Key:   "storage.oss.bucket",
				Value: "demo-bucket",
			},
			"storage.oss.publicBaseURL": {
				Key:   "storage.oss.publicBaseURL",
				Value: "https://cdn.example.com",
			},
			"storage.oss.accessKeyID": {
				Key:   "storage.oss.accessKeyID",
				Value: "ak",
			},
			"storage.oss.accessKeySecret": {
				Key:   "storage.oss.accessKeySecret",
				Value: "sk",
			},
		},
	}, config.TurnstileConfig{}, config.StorageConfig{Provider: "aliyun_oss"}, nil)

	svc := NewService(newMemoryRepo(), config.StorageConfig{Provider: "aliyun_oss"}, sysCfg, nil)
	got, ok := svc.storedPathFromPublicURL(context.Background(), " `https://cdn.example.com/blog/cache/draft-image.png` ")
	if !ok {
		t.Fatalf("storedPathFromPublicURL() ok = false, want true")
	}
	want := "oss:blog/cache/draft-image.png"
	if got != want {
		t.Fatalf("storedPathFromPublicURL() = %q, want %q", got, want)
	}
}

func TestPublicURLUsesSysConfigOSSSettings(t *testing.T) {
	t.Parallel()

	sysCfg := sysconfig.NewService(&fakeSysConfigStore{
		items: map[string]domainconfig.SysConfig{
			"storage.provider": {
				Key:   "storage.provider",
				Value: "aliyun_oss",
			},
			"storage.oss.region": {
				Key:   "storage.oss.region",
				Value: "cn-beijing",
			},
			"storage.oss.bucket": {
				Key:   "storage.oss.bucket",
				Value: "demo-bucket",
			},
			"storage.oss.prefix": {
				Key:   "storage.oss.prefix",
				Value: "blog-assets",
			},
			"storage.oss.publicBaseURL": {
				Key:   "storage.oss.publicBaseURL",
				Value: "https://cdn.example.com",
			},
			"storage.oss.accessKeyID": {
				Key:   "storage.oss.accessKeyID",
				Value: "ak",
			},
			"storage.oss.accessKeySecret": {
				Key:   "storage.oss.accessKeySecret",
				Value: "sk",
			},
		},
	}, config.TurnstileConfig{}, config.StorageConfig{Provider: "local"}, nil)

	svc := NewService(newMemoryRepo(), config.StorageConfig{}, sysCfg, nil)
	got := svc.PublicURL("oss:blog/images/cover.png")
	want := "https://cdn.example.com/blog-assets/blog/images/cover.png"
	if got != want {
		t.Fatalf("PublicURL() = %q, want %q", got, want)
	}
}

func TestResolveDuplicateHashUploadReusesExistingRecord(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   1,
		Name: "背景.jpg",
		Path: "/blog/cache/existing.jpg",
		Type: "picture",
		Size: 123,
		Hash: "same-hash",
	})
	svc := NewService(repo, config.StorageConfig{UploadDir: t.TempDir()}, nil, nil)

	existing, handled, err := svc.resolveDuplicateHashUpload(
		context.Background(),
		errors.New(`ERROR: duplicate key value violates unique constraint "uq_upload_file_hash" (SQLSTATE 23505)`),
		"same-hash",
		uploadSelection{kind: "picture", dir: "blog/cache"},
		svc.localStorage,
		"/blog/cache/new-file.jpg",
	)
	if !handled {
		t.Fatalf("resolveDuplicateHashUpload() handled = false, want true")
	}
	if err != nil {
		t.Fatalf("resolveDuplicateHashUpload() error = %v", err)
	}
	if existing == nil {
		t.Fatalf("resolveDuplicateHashUpload() existing = nil, want record")
	}
	if existing.ID != 1 {
		t.Fatalf("resolveDuplicateHashUpload() existing.ID = %d, want 1", existing.ID)
	}
}

func TestPrepareExistingUploadForTargetReturnsFinalRecordForCacheUpload(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   1,
		Name: "existing.jpg",
		Path: "/blog/images/existing.jpg",
		Type: "picture",
		Size: 123,
		Hash: "same-hash",
	})
	svc := NewService(repo, config.StorageConfig{UploadDir: t.TempDir()}, nil, nil)
	existing, err := repo.FindByHash(context.Background(), "same-hash")
	if err != nil {
		t.Fatalf("FindByHash() error = %v", err)
	}
	got, err := svc.prepareExistingUploadForTarget(
		context.Background(),
		existing,
		uploadSelection{kind: "picture", dir: "blog/cache"},
		svc.currentStorage(context.Background()),
	)
	if err != nil {
		t.Fatalf("prepareExistingUploadForTarget() error = %v", err)
	}
	if got.Path != "/blog/images/existing.jpg" {
		t.Fatalf("prepareExistingUploadForTarget() path = %q, want /blog/images/existing.jpg", got.Path)
	}
}

func TestPrepareExistingUploadForTargetPromotesCacheRecordForFinalUpload(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()
	cacheDiskPath := filepath.Join(uploadDir, "blog", "cache", "existing.jpg")
	writePNG(t, cacheDiskPath)

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   1,
		Name: "existing.jpg",
		Path: "/blog/cache/existing.jpg",
		Type: "picture",
		Size: 123,
		Hash: "same-hash",
	})
	svc := NewService(repo, config.StorageConfig{UploadDir: uploadDir}, nil, nil)
	existing, err := repo.FindByHash(context.Background(), "same-hash")
	if err != nil {
		t.Fatalf("FindByHash() error = %v", err)
	}
	got, err := svc.prepareExistingUploadForTarget(
		context.Background(),
		existing,
		uploadSelection{kind: "picture", dir: "blog/images"},
		svc.currentStorage(context.Background()),
	)
	if err != nil {
		t.Fatalf("prepareExistingUploadForTarget() error = %v", err)
	}
	if got.Path != "/blog/images/existing.jpg" {
		t.Fatalf("prepareExistingUploadForTarget() path = %q, want /blog/images/existing.jpg", got.Path)
	}
	if _, err := os.Stat(cacheDiskPath); !os.IsNotExist(err) {
		t.Fatalf("expected cache file to be moved, stat err = %v", err)
	}
	finalDiskPath := filepath.Join(uploadDir, "blog", "images", "existing.jpg")
	if _, err := os.Stat(finalDiskPath); err != nil {
		t.Fatalf("expected promoted file at %q: %v", finalDiskPath, err)
	}
	updated, err := repo.FindByPath(context.Background(), "/blog/images/existing.jpg")
	if err != nil {
		t.Fatalf("FindByPath(promoted) error = %v", err)
	}
	if updated.ID != 1 {
		t.Fatalf("promoted file.ID = %d, want 1", updated.ID)
	}
}

func TestPromoteDraftURLReturnsFinalURLWhenCacheRecordAlreadyPromoted(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   1,
		Name: "existing.jpg",
		Path: "/blog/images/existing.jpg",
		Type: "picture",
		Size: 123,
		Hash: "same-hash",
	})
	svc := NewService(repo, config.StorageConfig{UploadDir: t.TempDir()}, nil, nil)

	got, err := svc.PromoteDraftURL(context.Background(), "/uploads/blog/cache/existing.jpg")
	if err != nil {
		t.Fatalf("PromoteDraftURL() error = %v", err)
	}
	if got != "/uploads/blog/images/existing.jpg" {
		t.Fatalf("PromoteDraftURL() = %q, want /uploads/blog/images/existing.jpg", got)
	}
}

type memoryRepo struct {
	nextID int64
	files  []domainmedia.UploadFile
}

type fakeSysConfigStore struct {
	items map[string]domainconfig.SysConfig
}

func (r *fakeSysConfigStore) GetByKey(_ context.Context, key string) (*domainconfig.SysConfig, error) {
	item, ok := r.items[key]
	if !ok {
		return nil, domainconfig.ErrSysConfigNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (r *fakeSysConfigStore) List(_ context.Context, keys []string) ([]domainconfig.SysConfig, error) {
	if len(keys) == 0 {
		items := make([]domainconfig.SysConfig, 0, len(r.items))
		for _, item := range r.items {
			items = append(items, item)
		}
		return items, nil
	}
	allow := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		allow[key] = struct{}{}
	}
	items := make([]domainconfig.SysConfig, 0, len(keys))
	for key, item := range r.items {
		if _, ok := allow[key]; ok {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeSysConfigStore) Upsert(_ context.Context, configs []domainconfig.SysConfig) error {
	if r.items == nil {
		r.items = make(map[string]domainconfig.SysConfig, len(configs))
	}
	for _, cfg := range configs {
		r.items[cfg.Key] = cfg
	}
	return nil
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{nextID: 1}
}

func (r *memoryRepo) FindByHash(_ context.Context, hash string) (*domainmedia.UploadFile, error) {
	for i := range r.files {
		if r.files[i].Hash == hash {
			file := r.files[i]
			return &file, nil
		}
	}
	return nil, domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) FindByID(_ context.Context, id int64) (*domainmedia.UploadFile, error) {
	for i := range r.files {
		if r.files[i].ID == id {
			file := r.files[i]
			return &file, nil
		}
	}
	return nil, domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) FindByPath(_ context.Context, uploadPath string) (*domainmedia.UploadFile, error) {
	for i := range r.files {
		if r.files[i].Path == uploadPath {
			file := r.files[i]
			return &file, nil
		}
	}
	return nil, domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) Create(_ context.Context, file *domainmedia.UploadFile) error {
	if file.ID == 0 {
		file.ID = r.nextID
		r.nextID++
	}
	r.files = append(r.files, *file)
	return nil
}

func (r *memoryRepo) Update(_ context.Context, file *domainmedia.UploadFile) error {
	for i := range r.files {
		if r.files[i].ID == file.ID {
			r.files[i] = *file
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) UpdatePath(_ context.Context, id int64, path string) error {
	for i := range r.files {
		if r.files[i].ID == id {
			r.files[i].Path = path
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) UpdateName(_ context.Context, id int64, name string) error {
	for i := range r.files {
		if r.files[i].ID == id {
			r.files[i].Name = name
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) List(_ context.Context, offset int, limit int) ([]domainmedia.UploadFile, int64, error) {
	if offset > len(r.files) {
		return nil, int64(len(r.files)), nil
	}
	end := offset + limit
	if end > len(r.files) {
		end = len(r.files)
	}
	items := append([]domainmedia.UploadFile(nil), r.files[offset:end]...)
	return items, int64(len(r.files)), nil
}

func (r *memoryRepo) ListAll(_ context.Context) ([]domainmedia.UploadFile, error) {
	return append([]domainmedia.UploadFile(nil), r.files...), nil
}

func (r *memoryRepo) DeleteByID(_ context.Context, id int64) error {
	for i := range r.files {
		if r.files[i].ID == id {
			r.files = append(r.files[:i], r.files[i+1:]...)
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) mustSeed(file domainmedia.UploadFile) {
	r.files = append(r.files, file)
	if file.ID >= r.nextID {
		r.nextID = file.ID + 1
	}
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", path, err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 80, B: 40, A: 255})
		}
	}

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("png.Encode(%q) error = %v", path, err)
	}
}

func writeText(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
