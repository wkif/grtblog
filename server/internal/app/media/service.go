package media

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	goexif "github.com/rwcarlsen/goexif/exif"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/media"
)

type Service struct {
	repo               media.Repository
	uploadDir          string
	events             appEvent.Bus
	localStorage       *localStorage
	defaultUploadStore string
	defaultOSS         config.OSSConfig
	sysCfg             *sysconfig.Service
	cacheDir           string
	imageDir           string
	videoDir           string
	fileDir            string
}

func NewService(repo media.Repository, cfg config.StorageConfig, sysCfg *sysconfig.Service, events appEvent.Bus) *Service {
	local := newLocalStorage(cfg.UploadDir)
	if events == nil {
		events = appEvent.NopBus{}
	}
	service := &Service{
		repo:               repo,
		uploadDir:          local.uploadDir,
		events:             events,
		localStorage:       local,
		defaultUploadStore: storageProviderLocal,
		defaultOSS:         cfg.OSS,
		sysCfg:             sysCfg,
		cacheDir:           cleanManagedDir(cfg.CacheDir, "blog/cache"),
		imageDir:           cleanManagedDir(cfg.ImageDir, "blog/images"),
		videoDir:           cleanManagedDir(cfg.VideoDir, "blog/video"),
		fileDir:            cleanManagedDir(cfg.FileDir, "blog/files"),
	}
	if provider := strings.ToLower(strings.TrimSpace(cfg.Provider)); provider != "" {
		service.defaultUploadStore = provider
	}
	return service
}

const thumbnailMaxWidth = 1200
const thumbnailDir = "thumbnails"
const thumbnailQuality = 82

var managedURLPattern = regexp.MustCompile("https?://[^\\s<>\"'`\\)\\]]+|/uploads/[^\\s<>\"'`\\)\\]]+")

// ImageMeta 图片元信息，上传图片时自动提取。
type ImageMeta struct {
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	DominantColor string `json:"dominantColor,omitempty"` // hex, e.g. "#a3b2c1"
}

type UploadResult struct {
	File         media.UploadFile
	Created      bool
	ThumbnailURL string     // 缩略图公开路径（仅 picture 类型）
	ImageMeta    *ImageMeta // 图片元信息（仅 picture 类型）
}

type SyncResult struct {
	Scanned           int
	Indexed           int
	Created           int
	Updated           int
	Deleted           int
	SkippedDuplicates int
}

type indexedDiskFile struct {
	Name string
	Path string
	Type string
	Size int64
	Hash string
}

type uploadSelection struct {
	kind string
	dir  string
}

type managedDirs struct {
	cacheDir string
	imageDir string
	videoDir string
	fileDir  string
}

type runtimeStorage struct {
	provider string
	local    *localStorage
	oss      *aliyunOSSStorage
	dirs     managedDirs
}

func (s *Service) currentStorage(ctx context.Context) runtimeStorage {
	settings := config.StorageConfig{
		Provider: s.defaultUploadStore,
		CacheDir: s.cacheDir,
		ImageDir: s.imageDir,
		VideoDir: s.videoDir,
		FileDir:  s.fileDir,
		OSS:      s.defaultOSS,
	}
	if s.sysCfg != nil {
		settings = s.sysCfg.StorageSettings(ctx)
	}
	current := runtimeStorage{
		provider: normalizeProvider(settings.Provider),
		local:    s.localStorage,
		dirs: managedDirs{
			cacheDir: cleanManagedDir(settings.CacheDir, s.cacheDir),
			imageDir: cleanManagedDir(settings.ImageDir, s.imageDir),
			videoDir: cleanManagedDir(settings.VideoDir, s.videoDir),
			fileDir:  cleanManagedDir(settings.FileDir, s.fileDir),
		},
	}
	if ossStorage, err := newAliyunOSSStorage(settings.OSS); err == nil {
		current.oss = ossStorage
	}
	return current
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case storageProviderAliyunOSS:
		return storageProviderAliyunOSS
	case "", storageProviderLocal:
		return storageProviderLocal
	default:
		return storageProviderLocal
	}
}

func (s *Service) defaultBackend(current runtimeStorage) (storageBackend, error) {
	switch current.provider {
	case "", storageProviderLocal:
		return current.local, nil
	case storageProviderAliyunOSS:
		if current.oss == nil {
			return nil, errors.New("aliyun oss storage is not configured")
		}
		return current.oss, nil
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", current.provider)
	}
}

func (s *Service) backendForStoredPath(current runtimeStorage, storedPath string) storageBackend {
	if isOSSStoredPath(storedPath) {
		return current.oss
	}
	return current.local
}

func (s *Service) buildStoredPath(current runtimeStorage, store storageBackend, dir string, ext string) string {
	filename := s.buildFilename(dir, ext)
	if current.oss != nil && store == current.oss {
		return makeOSSStoredPath(path.Join(dir, filename))
	}
	return normalizeStoredPath("/" + dir + "/" + filename)
}

func (s *Service) Upload(ctx context.Context, file *multipart.FileHeader, fileType string) (*UploadResult, error) {
	if file == nil {
		return nil, errors.New("file is required")
	}

	current := s.currentStorage(ctx)
	selection, err := s.selectUploadTarget(fileType, file, current.dirs)
	if err != nil {
		return nil, err
	}

	hash, err := hashFile(file)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByHash(ctx, hash)
	if err != nil && !errors.Is(err, media.ErrUploadFileNotFound) {
		return nil, err
	}
	if existing != nil && existing.Type == selection.kind {
		existing, err = s.prepareExistingUploadForTarget(ctx, existing, selection, current)
		if err != nil {
			return nil, err
		}
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if existing != nil {
		if !isOSSStoredPath(existing.Path) {
			existingDisk := s.localStorage.diskPathFromStored(existing.Path)
			if fileExists(existingDisk) {
				thumbURL, meta := s.processImage(existing.Path, existing.Type)
				return &UploadResult{File: *existing, Created: false, ThumbnailURL: thumbURL, ImageMeta: meta}, nil
			}
			if err := s.localStorage.Upload(ctx, existing.Path, file); err != nil {
				return nil, err
			}
		}
		thumbURL, meta := s.processImage(existing.Path, existing.Type)
		return &UploadResult{File: *existing, Created: false, ThumbnailURL: thumbURL, ImageMeta: meta}, nil
	}

	store, err := s.defaultBackend(current)
	if err != nil {
		return nil, err
	}
	storedPath := s.buildStoredPath(current, store, selection.dir, ext)
	if err := store.Upload(ctx, storedPath, file); err != nil {
		return nil, err
	}

	record := &media.UploadFile{
		Name: file.Filename,
		Path: storedPath,
		Type: selection.kind,
		Size: file.Size,
		Hash: hash,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		if existing, handled, resolveErr := s.resolveDuplicateHashUpload(ctx, err, hash, selection, store, storedPath); handled {
			if resolveErr != nil {
				return nil, resolveErr
			}
			thumbURL, meta := s.processImage(existing.Path, existing.Type)
			return &UploadResult{File: *existing, Created: false, ThumbnailURL: thumbURL, ImageMeta: meta}, nil
		}
		return nil, err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "media.uploaded",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":   record.ID,
			"Name": record.Name,
			"Path": record.Path,
			"Type": record.Type,
			"Size": record.Size,
		},
	})
	thumbURL, meta := s.processImage(storedPath, selection.kind)
	return &UploadResult{File: *record, Created: true, ThumbnailURL: thumbURL, ImageMeta: meta}, nil
}

func (s *Service) resolveDuplicateHashUpload(
	ctx context.Context,
	createErr error,
	hash string,
	selection uploadSelection,
	store storageBackend,
	storedPath string,
) (*media.UploadFile, bool, error) {
	if !isDuplicateUploadHashError(createErr) {
		return nil, false, nil
	}
	if store != nil && storedPath != "" {
		_ = store.Delete(ctx, storedPath)
	}
	existing, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, true, err
	}
	if existing.Type != selection.kind {
		return nil, true, createErr
	}
	current := s.currentStorage(ctx)
	existing, err = s.prepareExistingUploadForTarget(ctx, existing, selection, current)
	if err != nil {
		return nil, true, err
	}
	return existing, true, nil
}

func isDuplicateUploadHashError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "uq_upload_file_hash") ||
		(strings.Contains(msg, "duplicate key value") && strings.Contains(msg, "upload_file")) ||
		strings.Contains(msg, "SQLSTATE 23505")
}

func (s *Service) prepareExistingUploadForTarget(
	ctx context.Context,
	existing *media.UploadFile,
	selection uploadSelection,
	current runtimeStorage,
) (*media.UploadFile, error) {
	if existing == nil {
		return nil, nil
	}
	targetIsCache := selection.dir == current.dirs.cacheDir
	existingIsCache := s.isCacheStoredPath(existing.Path, current.dirs)
	if targetIsCache || !existingIsCache {
		return existing, nil
	}
	targetDir := s.finalDirForType(existing.Type, current.dirs)
	return s.promoteRecordToDir(ctx, existing, targetDir, current)
}

func (s *Service) promoteRecordToDir(
	ctx context.Context,
	record *media.UploadFile,
	targetDir string,
	current runtimeStorage,
) (*media.UploadFile, error) {
	targetStoredPath := s.buildPromotedStoredPath(record.Path, targetDir)
	if targetStoredPath == "" || targetStoredPath == record.Path {
		return record, nil
	}
	switch backend := s.backendForStoredPath(current, record.Path).(type) {
	case *localStorage:
		if err := backend.Move(record.Path, targetStoredPath); err != nil {
			return nil, err
		}
	case *aliyunOSSStorage:
		if err := backend.Move(ctx, record.Path, targetStoredPath); err != nil {
			return nil, err
		}
	case nil:
		return nil, errors.New("storage backend is not configured")
	default:
		return nil, errors.New("unsupported storage backend for draft promotion")
	}
	record.Path = targetStoredPath
	record.Type = s.detectTypeByPath(record.Path)
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

type ListResult struct {
	Items []media.UploadFile
	Total int64
	Page  int
	Size  int
}

func (s *Service) List(ctx context.Context, page int, size int) (*ListResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size
	items, total, err := s.repo.List(ctx, offset, size)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *Service) SyncIndex(ctx context.Context) (*SyncResult, error) {
	existing, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	diskFiles, scanned, skippedDuplicates, err := s.scanUploadFiles()
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		Scanned:           scanned,
		Indexed:           len(diskFiles),
		SkippedDuplicates: skippedDuplicates,
	}

	existingByPath := make(map[string]*media.UploadFile, len(existing))
	existingByHash := make(map[string]*media.UploadFile, len(existing))
	for i := range existing {
		file := &existing[i]
		if isOSSStoredPath(file.Path) {
			continue
		}
		existingByPath[file.Path] = file
		if strings.TrimSpace(file.Hash) != "" {
			existingByHash[file.Hash] = file
		}
	}

	usedIDs := make(map[int64]struct{}, len(diskFiles))
	for _, diskFile := range diskFiles {
		pathRecord := existingByPath[diskFile.Path]
		hashRecord := existingByHash[diskFile.Hash]

		target := resolveSyncTarget(pathRecord, hashRecord)
		if target == nil {
			record := &media.UploadFile{
				Name: diskFile.Name,
				Path: diskFile.Path,
				Type: diskFile.Type,
				Size: diskFile.Size,
				Hash: diskFile.Hash,
			}
			if err := s.repo.Create(ctx, record); err != nil {
				return nil, err
			}
			result.Created++
			usedIDs[record.ID] = struct{}{}
			existingByPath[record.Path] = record
			existingByHash[record.Hash] = record
			continue
		}

		usedIDs[target.ID] = struct{}{}
		if !needsSyncUpdate(*target, diskFile) {
			continue
		}

		originalPath := target.Path
		originalHash := target.Hash
		target.Name = diskFile.Name
		target.Path = diskFile.Path
		target.Type = diskFile.Type
		target.Size = diskFile.Size
		target.Hash = diskFile.Hash
		if err := s.repo.Update(ctx, target); err != nil {
			return nil, err
		}
		result.Updated++
		if originalPath != target.Path {
			delete(existingByPath, originalPath)
		}
		existingByPath[target.Path] = target
		if strings.TrimSpace(originalHash) != "" && originalHash != target.Hash {
			delete(existingByHash, originalHash)
		}
		existingByHash[target.Hash] = target
	}

	for i := range existing {
		file := &existing[i]
		if isOSSStoredPath(file.Path) {
			continue
		}
		if _, ok := usedIDs[file.ID]; ok {
			continue
		}
		if err := s.repo.DeleteByID(ctx, file.ID); err != nil {
			return nil, err
		}
		result.Deleted++
	}

	return result, nil
}

func (s *Service) Rename(ctx context.Context, id int64, name string) (*media.UploadFile, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, errors.New("name is required")
	}
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if file.Name == trimmed {
		return file, nil
	}
	if err := s.repo.UpdateName(ctx, id, trimmed); err != nil {
		return nil, err
	}
	file.Name = trimmed
	return file, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (*media.UploadFile, error) {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	current := s.currentStorage(ctx)
	backend := s.backendForStoredPath(current, file.Path)
	if backend == nil {
		return nil, errors.New("storage backend is not configured")
	}
	if err := backend.Delete(ctx, file.Path); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "media.deleted",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":   file.ID,
			"Name": file.Name,
			"Path": file.Path,
			"Type": file.Type,
			"Size": file.Size,
		},
	})
	return file, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*media.UploadFile, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) PublicURL(storedPath string) string {
	current := s.currentStorage(context.Background())
	backend := s.backendForStoredPath(current, storedPath)
	if backend == nil {
		return ""
	}
	return backend.PublicURL(storedPath)
}

func (s *Service) OpenStoredFile(ctx context.Context, storedPath string) (io.ReadCloser, error) {
	current := s.currentStorage(ctx)
	backend := s.backendForStoredPath(current, storedPath)
	if backend == nil {
		return nil, errors.New("storage backend is not configured")
	}
	return backend.Open(ctx, storedPath)
}

func (s *Service) PromoteDraftURL(ctx context.Context, publicURL string) (string, error) {
	publicURL = normalizeManagedPublicURLInput(publicURL)
	if publicURL == "" {
		return publicURL, nil
	}
	current := s.currentStorage(ctx)
	storedPath, ok := s.storedPathFromPublicURL(ctx, publicURL)
	if !ok || !s.isCacheStoredPath(storedPath, current.dirs) {
		return publicURL, nil
	}
	record, err := s.repo.FindByPath(ctx, storedPath)
	if err != nil {
		if errors.Is(err, media.ErrUploadFileNotFound) {
			if promoted := s.findPromotedRecordForCachePath(ctx, storedPath, current.dirs); promoted != nil {
				return s.PublicURL(promoted.Path), nil
			}
		}
		return "", err
	}
	targetDir := s.finalDirForType(record.Type, current.dirs)
	targetStoredPath := s.buildPromotedStoredPath(record.Path, targetDir)
	if targetStoredPath == "" || targetStoredPath == record.Path {
		return s.PublicURL(record.Path), nil
	}

	switch backend := s.backendForStoredPath(current, record.Path).(type) {
	case *localStorage:
		if err := backend.Move(record.Path, targetStoredPath); err != nil {
			return "", err
		}
	case *aliyunOSSStorage:
		if err := backend.Move(ctx, record.Path, targetStoredPath); err != nil {
			return "", err
		}
	case nil:
		return "", errors.New("storage backend is not configured")
	default:
		return "", errors.New("unsupported storage backend for draft promotion")
	}

	record.Path = targetStoredPath
	record.Type = s.detectTypeByPath(record.Path)
	if err := s.repo.Update(ctx, record); err != nil {
		return "", err
	}
	return s.PublicURL(record.Path), nil
}

func (s *Service) findPromotedRecordForCachePath(ctx context.Context, storedPath string, dirs managedDirs) *media.UploadFile {
	if !s.isCacheStoredPath(storedPath, dirs) {
		return nil
	}
	fileType := s.detectTypeByPath(storedPath)
	candidates := []string{
		s.buildPromotedStoredPath(storedPath, s.finalDirForType(fileType, dirs)),
	}
	for _, dir := range []string{dirs.imageDir, dirs.videoDir, dirs.fileDir} {
		candidate := s.buildPromotedStoredPath(storedPath, dir)
		if candidate == "" {
			continue
		}
		seen := false
		for _, existing := range candidates {
			if existing == candidate {
				seen = true
				break
			}
		}
		if !seen {
			candidates = append(candidates, candidate)
		}
	}
	for _, candidate := range candidates {
		record, err := s.repo.FindByPath(ctx, candidate)
		if err == nil {
			return record
		}
		if err != nil && !errors.Is(err, media.ErrUploadFileNotFound) {
			return nil
		}
	}
	return nil
}

func (s *Service) RewriteDraftURLs(ctx context.Context, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}
	matches := managedURLPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return text, nil
	}
	replacements := make([]string, 0, len(matches)*2)
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if _, ok := seen[match]; ok {
			continue
		}
		seen[match] = struct{}{}
		rewritten, err := s.PromoteDraftURL(ctx, match)
		if err != nil {
			return "", err
		}
		if rewritten != match {
			replacements = append(replacements, match, rewritten)
		}
	}
	if len(replacements) == 0 {
		return text, nil
	}
	return strings.NewReplacer(replacements...).Replace(text), nil
}

func (s *Service) ResolveDiskPath(storedPath string) (string, error) {
	if isOSSStoredPath(storedPath) {
		return "", errors.New("oss stored file does not have a local disk path")
	}
	diskPath := s.localStorage.diskPathFromStored(storedPath)
	if diskPath == "" {
		return "", errors.New("empty stored path")
	}
	return diskPath, nil
}

// processImage 为图片生成缩略图并提取元信息（尺寸 + 主色调）。
func (s *Service) processImage(storedPath string, fileType string) (thumbURL string, meta *ImageMeta) {
	if strings.ToLower(strings.TrimSpace(fileType)) != "picture" {
		return "", nil
	}
	if isOSSStoredPath(storedPath) {
		reader, err := s.OpenStoredFile(context.Background(), storedPath)
		if err != nil {
			log.Printf("[image] open failed for %s: %v", storedPath, err)
			return "", nil
		}
		defer reader.Close()
		src, _, err := image.Decode(reader)
		if err != nil {
			log.Printf("[image] decode failed for %s: %v", storedPath, err)
			return "", nil
		}
		bounds := src.Bounds()
		return "", &ImageMeta{
			Width:         bounds.Dx(),
			Height:        bounds.Dy(),
			DominantColor: calcDominantColor(src),
		}
	}

	diskPath := s.localStorage.diskPathFromStored(storedPath)
	f, err := os.Open(diskPath)
	if err != nil {
		log.Printf("[image] open failed for %s: %v", diskPath, err)
		return "", nil
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		log.Printf("[image] decode failed for %s: %v", diskPath, err)
		return "", nil
	}

	bounds := src.Bounds()
	meta = &ImageMeta{
		Width:         bounds.Dx(),
		Height:        bounds.Dy(),
		DominantColor: calcDominantColor(src),
	}

	// Generate thumbnail
	thumbStoredPath := "/" + thumbnailDir + storedPath
	thumbDiskPath := s.localStorage.diskPathFromStored(thumbStoredPath)

	if !fileExists(thumbDiskPath) {
		thumb := imaging.Resize(src, thumbnailMaxWidth, 0, imaging.Lanczos)
		if err := os.MkdirAll(filepath.Dir(thumbDiskPath), 0o755); err != nil {
			log.Printf("[thumbnail] mkdir failed: %v", err)
			return "", meta
		}
		out, err := os.Create(thumbDiskPath)
		if err != nil {
			log.Printf("[thumbnail] create failed: %v", err)
			return "", meta
		}
		defer out.Close()
		if err := jpeg.Encode(out, thumb, &jpeg.Options{Quality: thumbnailQuality}); err != nil {
			log.Printf("[thumbnail] encode failed: %v", err)
			return "", meta
		}
	}

	return s.localStorage.PublicURL(thumbStoredPath), meta
}

// calcDominantColor 采样缩小后取平均色。
func calcDominantColor(img image.Image) string {
	small := imaging.Resize(img, 32, 0, imaging.Box)
	bounds := small.Bounds()
	var r, g, b uint64
	var count uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			cr, cg, cb, ca := small.At(x, y).RGBA()
			if ca < 0x1000 {
				continue
			}
			r += uint64(cr >> 8)
			g += uint64(cg >> 8)
			b += uint64(cb >> 8)
			count++
		}
	}
	if count == 0 {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", r/count, g/count, b/count)
}

// ThumbnailURLFor 根据原图公开 URL 返回对应缩略图的公开 URL。
// 如果缩略图不存在于磁盘，返回空字符串。
func (s *Service) ThumbnailURLFor(publicURL string) string {
	storedPath, ok := s.storedPathFromPublicURL(context.Background(), publicURL)
	if !ok || isOSSStoredPath(storedPath) {
		return ""
	}
	thumbStoredPath := "/" + thumbnailDir + storedPath
	thumbDiskPath := s.localStorage.diskPathFromStored(thumbStoredPath)
	if fileExists(thumbDiskPath) {
		return s.localStorage.PublicURL(thumbStoredPath)
	}
	return ""
}

// ExtractImageMetaFromURL 根据本站公开 URL 提取图片元信息（尺寸+主色调）并确保缩略图存在。
// 外链返回 nil。
func (s *Service) ExtractImageMetaFromURL(publicURL string) (thumbURL string, meta *ImageMeta) {
	thumbURL, meta, _ = s.ExtractPhotoMetadataFromURL(publicURL)
	return thumbURL, meta
}

// ExtractPhotoMetadataFromURL 根据本站公开 URL 提取图片元信息和 EXIF 摘要。
// 外链或不存在的本地文件返回空结果。
func (s *Service) ExtractPhotoMetadataFromURL(publicURL string) (thumbURL string, meta *ImageMeta, exifData map[string]any) {
	storedPath, ok := s.storedPathFromPublicURL(context.Background(), publicURL)
	if !ok {
		return "", nil, nil
	}
	if isOSSStoredPath(storedPath) {
		thumbURL, meta = s.processImage(storedPath, "picture")
		return thumbURL, meta, nil
	}
	diskPath := s.localStorage.diskPathFromStored(storedPath)
	if !fileExists(diskPath) {
		return "", nil, nil
	}
	thumbURL, meta = s.processImage(storedPath, "picture")
	exifData = extractExifSummary(diskPath)
	return thumbURL, meta, exifData
}

func (s *Service) storedPathFromPublicURL(ctx context.Context, publicURL string) (string, bool) {
	publicURL = normalizeManagedPublicURLInput(publicURL)
	if publicURL == "" {
		return "", false
	}
	const localPrefix = "/uploads"
	if strings.HasPrefix(publicURL, localPrefix) {
		storedPath := normalizeStoredPath(strings.TrimPrefix(publicURL, localPrefix))
		return storedPath, storedPath != ""
	}
	current := s.currentStorage(ctx)
	if current.oss != nil {
		return current.oss.matchPublicURL(publicURL)
	}
	return "", false
}

func normalizeManagedPublicURLInput(value string) string {
	trimmed := strings.TrimSpace(value)
	for {
		next := strings.Trim(trimmed, "`\"'<>")
		next = strings.TrimSpace(next)
		if next == trimmed {
			return next
		}
		trimmed = next
	}
}

func (s *Service) selectUploadTarget(fileType string, file *multipart.FileHeader, dirs managedDirs) (uploadSelection, error) {
	requested := strings.ToLower(strings.TrimSpace(fileType))
	switch requested {
	case "picture":
		return uploadSelection{kind: "picture", dir: dirs.imageDir}, nil
	case "video":
		return uploadSelection{kind: "video", dir: dirs.videoDir}, nil
	case "file":
		return uploadSelection{kind: "file", dir: dirs.fileDir}, nil
	case "cache":
		kind := detectUploadKind(file)
		return uploadSelection{kind: kind, dir: dirs.cacheDir}, nil
	default:
		return uploadSelection{}, media.ErrInvalidUploadType
	}
}

func (s *Service) isCacheStoredPath(storedPath string, dirs managedDirs) bool {
	return s.dirMatchesStoredPath(storedPath, dirs.cacheDir)
}

func (s *Service) finalDirForType(fileType string, dirs managedDirs) string {
	switch strings.ToLower(strings.TrimSpace(fileType)) {
	case "picture":
		return dirs.imageDir
	case "video":
		return dirs.videoDir
	default:
		return dirs.fileDir
	}
}

func (s *Service) buildPromotedStoredPath(storedPath string, dir string) string {
	baseName := path.Base(strings.TrimPrefix(strings.TrimPrefix(normalizeStoredPath(storedPath), ossStoredPathPrefix), "/"))
	if baseName == "" || baseName == "." {
		return ""
	}
	if isOSSStoredPath(storedPath) {
		return makeOSSStoredPath(path.Join(dir, baseName))
	}
	return normalizeStoredPath("/" + path.Join(dir, baseName))
}

func (s *Service) dirMatchesStoredPath(storedPath string, dir string) bool {
	candidate := path.Clean("/" + strings.Trim(dir, "/"))
	if candidate == "/" {
		return false
	}
	normalized := normalizeStoredPath(storedPath)
	if isOSSStoredPath(normalized) {
		return strings.HasPrefix(objectKeyFromStoredPath(normalized), strings.TrimPrefix(candidate, "/")+"/") ||
			objectKeyFromStoredPath(normalized) == strings.TrimPrefix(candidate, "/")
	}
	return strings.HasPrefix(normalized, candidate+"/") || normalized == candidate
}

func resolveSyncTarget(pathRecord *media.UploadFile, hashRecord *media.UploadFile) *media.UploadFile {
	switch {
	case pathRecord != nil && hashRecord != nil:
		if pathRecord.ID == hashRecord.ID {
			return pathRecord
		}
		return hashRecord
	case pathRecord != nil:
		return pathRecord
	default:
		return hashRecord
	}
}

func needsSyncUpdate(record media.UploadFile, diskFile indexedDiskFile) bool {
	return record.Name != diskFile.Name ||
		record.Path != diskFile.Path ||
		record.Type != diskFile.Type ||
		record.Size != diskFile.Size ||
		record.Hash != diskFile.Hash
}

func (s *Service) scanUploadFiles() ([]indexedDiskFile, int, int, error) {
	root := filepath.Clean(s.uploadDir)
	candidatesByHash := make(map[string]indexedDiskFile)
	scanned := 0
	skippedDuplicates := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			if s.shouldSkipSyncDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		scanned++
		hash, err := hashDiskPath(path)
		if err != nil {
			return err
		}
		if _, exists := candidatesByHash[hash]; exists {
			skippedDuplicates++
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}

		storedPath := "/" + filepath.ToSlash(relPath)
		candidatesByHash[hash] = indexedDiskFile{
			Name: filepath.Base(path),
			Path: storedPath,
			Type: detectIndexedFileType(storedPath),
			Size: info.Size(),
			Hash: hash,
		}
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}

	candidates := make([]indexedDiskFile, 0, len(candidatesByHash))
	for _, file := range candidatesByHash {
		candidates = append(candidates, file)
	}
	slices.SortFunc(candidates, func(a, b indexedDiskFile) int {
		return strings.Compare(a.Path, b.Path)
	})
	return candidates, scanned, skippedDuplicates, nil
}

func (s *Service) shouldSkipSyncDir(path string) bool {
	root := filepath.Clean(s.uploadDir)
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	first := filepath.ToSlash(rel)
	if idx := strings.IndexByte(first, '/'); idx >= 0 {
		first = first[:idx]
	}
	return first == thumbnailDir
}

func detectIndexedFileType(storedPath string) string {
	clean := strings.ToLower(filepath.ToSlash(strings.TrimSpace(storedPath)))
	if strings.HasPrefix(clean, "/blog/images/") || strings.HasPrefix(clean, "/pictures/") {
		return "picture"
	}
	if strings.HasPrefix(clean, "/blog/video/") {
		return "video"
	}

	switch strings.ToLower(filepath.Ext(clean)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".avif", ".heic", ".heif", ".tif", ".tiff":
		return "picture"
	case ".mp4", ".mov", ".m4v", ".webm", ".mkv", ".avi", ".wmv", ".flv", ".mpeg", ".mpg":
		return "video"
	default:
		return "file"
	}
}

func (s *Service) buildFilename(dir string, ext string) string {
	base := time.Now().Format("2006-01-02-15:04:05")
	ext = strings.TrimSpace(ext)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	for i := 0; i < 5; i++ {
		suffix := randomHex(2)
		filename := base + "-" + suffix + ext
		if !fileExists(s.localStorage.diskPathFromStored("/" + dir + "/" + filename)) {
			return filename
		}
	}
	suffix := randomHex(4)
	return base + "-" + suffix + ext
}

func randomHex(n int) string {
	if n <= 0 {
		return ""
	}
	byteLen := (n + 1) / 2
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)[:n]
	}
	fallback := hex.EncodeToString([]byte(time.Now().Format("150405.000")))
	if len(fallback) >= n {
		return fallback[:n]
	}
	return fallback
}

func cleanManagedDir(value string, fallback string) string {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func detectUploadKind(file *multipart.FileHeader) string {
	if file == nil {
		return "file"
	}
	contentType := strings.ToLower(strings.TrimSpace(file.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "image/") {
		return "picture"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	return detectTypeByName(file.Filename)
}

func detectTypeByName(name string) string {
	switch strings.ToLower(strings.TrimSpace(filepath.Ext(name))) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".avif", ".heic", ".heif", ".tif", ".tiff":
		return "picture"
	case ".mp4", ".mov", ".m4v", ".webm", ".mkv", ".avi", ".wmv", ".flv", ".mpeg", ".mpg":
		return "video"
	default:
		return "file"
	}
}

func (s *Service) detectTypeByPath(storedPath string) string {
	normalized := normalizeStoredPath(storedPath)
	if isOSSStoredPath(normalized) {
		return detectIndexedFileType("/" + objectKeyFromStoredPath(normalized))
	}
	return detectIndexedFileType(normalized)
}

func hashFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashDiskPath(path string) (string, error) {
	src, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func extractExifSummary(diskPath string) map[string]any {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	x, err := goexif.Decode(file)
	if err != nil {
		return nil
	}

	result := map[string]any{}
	if makeValue := exifString(x, goexif.Make); makeValue != "" {
		result["make"] = makeValue
	}
	if modelValue := exifString(x, goexif.Model); modelValue != "" {
		result["model"] = modelValue
	}
	if lensModel := exifString(x, goexif.LensModel); lensModel != "" {
		result["lensModel"] = lensModel
	}
	if focalLength := exifFocalLength(x); focalLength != "" {
		result["focalLength"] = focalLength
	}
	if fNumber := exifDecimal(x, goexif.FNumber, 2); fNumber != "" {
		result["fNumber"] = fNumber
	}
	if exposureTime := exifExposureTime(x); exposureTime != "" {
		result["exposureTime"] = exposureTime
	}
	if iso, ok := exifInt(x, goexif.ISOSpeedRatings); ok && iso > 0 {
		result["iso"] = iso
	}
	if dateTime, ok := exifDateTime(x); ok {
		result["dateTimeOriginal"] = dateTime
	}
	if lat, long, ok := exifLatLong(x); ok {
		result["gpsLatitude"] = lat
		result["gpsLongitude"] = long
	}
	if altitude, ok := exifAltitude(x); ok {
		result["gpsAltitude"] = altitude
	}
	if width, ok := exifInt(x, goexif.PixelXDimension); ok && width > 0 {
		result["imageWidth"] = width
	}
	if height, ok := exifInt(x, goexif.PixelYDimension); ok && height > 0 {
		result["imageHeight"] = height
	}
	if orientation, ok := exifInt(x, goexif.Orientation); ok && orientation > 0 {
		result["orientation"] = orientation
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func exifString(x *goexif.Exif, field goexif.FieldName) string {
	tag, err := x.Get(field)
	if err != nil {
		return ""
	}
	value, err := tag.StringVal()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func exifInt(x *goexif.Exif, field goexif.FieldName) (int, bool) {
	tag, err := x.Get(field)
	if err != nil {
		return 0, false
	}
	value, err := tag.Int(0)
	if err != nil {
		return 0, false
	}
	return value, true
}

func exifDecimal(x *goexif.Exif, field goexif.FieldName, precision int) string {
	tag, err := x.Get(field)
	if err != nil {
		return ""
	}
	value, ok := exifRat(tag)
	if !ok {
		return ""
	}
	return formatExifFloat(value, precision)
}

func exifFocalLength(x *goexif.Exif) string {
	tag, err := x.Get(goexif.FocalLength)
	if err != nil {
		return ""
	}
	value, ok := exifRat(tag)
	if !ok {
		return ""
	}
	return formatExifFloat(value, 2) + " mm"
}

func exifExposureTime(x *goexif.Exif) string {
	tag, err := x.Get(goexif.ExposureTime)
	if err != nil {
		return ""
	}
	num, den, err := tag.Rat2(0)
	if err != nil || den == 0 {
		return ""
	}
	if num > 0 && num < den {
		return fmt.Sprintf("%d/%d", num, den)
	}
	return formatExifFloat(float64(num)/float64(den), 4)
}

func exifDateTime(x *goexif.Exif) (string, bool) {
	tm, err := x.DateTime()
	if err != nil || tm.IsZero() {
		return "", false
	}
	return tm.Format("2006:01:02 15:04:05"), true
}

func exifLatLong(x *goexif.Exif) (float64, float64, bool) {
	lat, long, err := x.LatLong()
	if err != nil {
		return 0, 0, false
	}
	return lat, long, true
}

func exifAltitude(x *goexif.Exif) (float64, bool) {
	tag, err := x.Get(goexif.GPSAltitude)
	if err != nil {
		return 0, false
	}
	value, ok := exifRat(tag)
	if !ok {
		return 0, false
	}
	if ref, ok := exifInt(x, goexif.GPSAltitudeRef); ok && ref == 1 {
		value = -value
	}
	return value, true
}

func exifRat(tag interface {
	Rat2(i int) (num, den int64, err error)
}) (float64, bool) {
	num, den, err := tag.Rat2(0)
	if err != nil || den == 0 {
		return 0, false
	}
	return float64(num) / float64(den), true
}

func formatExifFloat(value float64, precision int) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return ""
	}
	if precision < 0 {
		precision = 0
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.*f", precision, value), "0"), ".")
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func removeFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
