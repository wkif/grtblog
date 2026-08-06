package sysconfig

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

const (
	uploadMaxSizeKey  = "upload.maxSizeMB"
	uploadCacheDirKey = "upload.cacheDir"
	uploadImageDirKey = "upload.imageDir"
	uploadVideoDirKey = "upload.videoDir"
	uploadFileDirKey  = "upload.fileDir"

	storageProviderKey           = "storage.provider"
	storageOSSRegionKey          = "storage.oss.region"
	storageOSSEndpointKey        = "storage.oss.endpoint"
	storageOSSBucketKey          = "storage.oss.bucket"
	storageOSSPrefixKey          = "storage.oss.prefix"
	storageOSSPublicBaseURLKey   = "storage.oss.publicBaseURL"
	storageOSSAccessKeyIDKey     = "storage.oss.accessKeyID"
	storageOSSAccessKeySecretKey = "storage.oss.accessKeySecret"
	storageOSSSecurityTokenKey   = "storage.oss.securityToken"
)

type UploadStorageSettings struct {
	CacheDir string
	ImageDir string
	VideoDir string
	FileDir  string
}

func (s *Service) UploadStorageSettings(ctx context.Context) UploadStorageSettings {
	settings := UploadStorageSettings{
		CacheDir: normalizeUploadManagedDir(s.defaultStorage.CacheDir, "blog/cache"),
		ImageDir: normalizeUploadManagedDir(s.defaultStorage.ImageDir, "blog/images"),
		VideoDir: normalizeUploadManagedDir(s.defaultStorage.VideoDir, "blog/video"),
		FileDir:  normalizeUploadManagedDir(s.defaultStorage.FileDir, "blog/files"),
	}
	apply := func(key string, target *string) {
		cfg, err := s.repo.GetByKey(ctx, key)
		if err != nil {
			return
		}
		if val := normalizeUploadManagedDir(cfg.Value, ""); val != "" {
			*target = val
		}
	}
	apply(uploadCacheDirKey, &settings.CacheDir)
	apply(uploadImageDirKey, &settings.ImageDir)
	apply(uploadVideoDirKey, &settings.VideoDir)
	apply(uploadFileDirKey, &settings.FileDir)
	return settings
}

func (s *Service) StorageSettings(ctx context.Context) config.StorageConfig {
	settings := s.defaultStorage
	settings.Provider = normalizeStorageProvider(settings.Provider)
	settings.CacheDir = normalizeUploadManagedDir(settings.CacheDir, "blog/cache")
	settings.ImageDir = normalizeUploadManagedDir(settings.ImageDir, "blog/images")
	settings.VideoDir = normalizeUploadManagedDir(settings.VideoDir, "blog/video")
	settings.FileDir = normalizeUploadManagedDir(settings.FileDir, "blog/files")
	settings.OSS.Prefix = strings.Trim(strings.TrimSpace(settings.OSS.Prefix), "/")
	settings.OSS.PublicBaseURL = strings.TrimRight(strings.TrimSpace(settings.OSS.PublicBaseURL), "/")

	uploadSettings := s.UploadStorageSettings(ctx)
	settings.CacheDir = uploadSettings.CacheDir
	settings.ImageDir = uploadSettings.ImageDir
	settings.VideoDir = uploadSettings.VideoDir
	settings.FileDir = uploadSettings.FileDir

	applyString := func(key string, apply func(string)) {
		cfg, err := s.repo.GetByKey(ctx, key)
		if err != nil {
			return
		}
		apply(strings.TrimSpace(cfg.Value))
	}

	applyString(storageProviderKey, func(value string) {
		if normalized := normalizeStorageProvider(value); normalized != "" {
			settings.Provider = normalized
		}
	})
	applyString(storageOSSRegionKey, func(value string) {
		settings.OSS.Region = value
	})
	applyString(storageOSSEndpointKey, func(value string) {
		settings.OSS.Endpoint = value
	})
	applyString(storageOSSBucketKey, func(value string) {
		settings.OSS.Bucket = value
	})
	applyString(storageOSSPrefixKey, func(value string) {
		settings.OSS.Prefix = strings.Trim(value, "/")
	})
	applyString(storageOSSPublicBaseURLKey, func(value string) {
		settings.OSS.PublicBaseURL = strings.TrimRight(value, "/")
	})
	applyString(storageOSSAccessKeyIDKey, func(value string) {
		settings.OSS.AccessKeyID = value
	})
	applyString(storageOSSAccessKeySecretKey, func(value string) {
		settings.OSS.AccessKeySecret = value
	})
	applyString(storageOSSSecurityTokenKey, func(value string) {
		settings.OSS.SecurityToken = value
	})

	return settings
}

func (s *Service) StorageSettingsWithOverrides(ctx context.Context, items []UpdateItem) (config.StorageConfig, error) {
	settings := s.StorageSettings(ctx)
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" || item.Value == nil {
			continue
		}

		builtin, ok := s.builtinConfigByKey(key)
		if !ok {
			continue
		}
		valueType := normalizeValueType(builtin.ValueType)
		parsed, isEmpty, err := parseValueByType(valueType, *item.Value)
		if err != nil {
			return settings, err
		}
		if builtin.IsSensitive && isEmpty {
			continue
		}
		if err := s.validateCustomValue(key, valueType, parsed); err != nil {
			return settings, err
		}

		switch key {
		case storageProviderKey:
			settings.Provider = normalizeStorageProvider(parsed)
		case uploadCacheDirKey:
			settings.CacheDir = normalizeUploadManagedDir(parsed, settings.CacheDir)
		case uploadImageDirKey:
			settings.ImageDir = normalizeUploadManagedDir(parsed, settings.ImageDir)
		case uploadVideoDirKey:
			settings.VideoDir = normalizeUploadManagedDir(parsed, settings.VideoDir)
		case uploadFileDirKey:
			settings.FileDir = normalizeUploadManagedDir(parsed, settings.FileDir)
		case storageOSSRegionKey:
			settings.OSS.Region = strings.TrimSpace(parsed)
		case storageOSSEndpointKey:
			settings.OSS.Endpoint = strings.TrimSpace(parsed)
		case storageOSSBucketKey:
			settings.OSS.Bucket = strings.TrimSpace(parsed)
		case storageOSSPrefixKey:
			settings.OSS.Prefix = strings.Trim(strings.TrimSpace(parsed), "/")
		case storageOSSPublicBaseURLKey:
			settings.OSS.PublicBaseURL = strings.TrimRight(strings.TrimSpace(parsed), "/")
		case storageOSSAccessKeyIDKey:
			settings.OSS.AccessKeyID = strings.TrimSpace(parsed)
		case storageOSSAccessKeySecretKey:
			settings.OSS.AccessKeySecret = strings.TrimSpace(parsed)
		case storageOSSSecurityTokenKey:
			settings.OSS.SecurityToken = strings.TrimSpace(parsed)
		}
	}
	return settings, nil
}

func (s *Service) builtinUploadConfigs() []domainconfig.SysConfig {
	return []domainconfig.SysConfig{
		newStringBuiltinConfig(uploadMaxSizeKey, "storage/upload", "上传大小限制", "上传文件最大大小，单位 MB；修改后需重启服务生效。", "50", 10),
		newStringBuiltinConfig(uploadCacheDirKey, "storage/upload", "草稿缓存目录", "编辑阶段上传资源的临时目录，保存时会转正。", normalizeUploadManagedDir(s.defaultStorage.CacheDir, "blog/cache"), 20),
		newStringBuiltinConfig(uploadImageDirKey, "storage/upload", "图片正式目录", "文章、页面、手记保存后，图片资源转入的目录。", normalizeUploadManagedDir(s.defaultStorage.ImageDir, "blog/images"), 30),
		newStringBuiltinConfig(uploadVideoDirKey, "storage/upload", "视频正式目录", "保存后视频资源转入的目录。", normalizeUploadManagedDir(s.defaultStorage.VideoDir, "blog/video"), 40),
		newStringBuiltinConfig(uploadFileDirKey, "storage/upload", "文件正式目录", "保存后普通附件转入的目录。", normalizeUploadManagedDir(s.defaultStorage.FileDir, "blog/files"), 50),
	}
}

func (s *Service) builtinStorageBackendConfigs() []domainconfig.SysConfig {
	ossVisibleWhen := mustRawJSON(`[{"key":"storage.provider","op":"eq","value":"aliyun_oss"}]`)
	providerOptions := mustRawJSON(`[
		{"label":"本地存储","value":"local"},
		{"label":"阿里云 OSS","value":"aliyun_oss"}
	]`)
	passwordMeta := mustRawJSON(`{"inputType":"password"}`)
	textareaMeta := mustRawJSON(`{"inputType":"textarea"}`)

	return []domainconfig.SysConfig{
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageProviderKey,
			GroupPath:    "storage/provider",
			Label:        "默认存储后端",
			Description:  "新上传文件默认写入的存储后端。历史文件仍按各自原始存储位置读取。",
			ValueType:    valueTypeEnum,
			DefaultValue: normalizeStorageProvider(s.defaultStorage.Provider),
			EnumOptions:  providerOptions,
			Sort:         10,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSRegionKey,
			GroupPath:    "storage/oss",
			Label:        "OSS Region",
			Description:  "阿里云 OSS Region，例如 cn-beijing。",
			DefaultValue: strings.TrimSpace(s.defaultStorage.OSS.Region),
			VisibleWhen:  ossVisibleWhen,
			Sort:         10,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSEndpointKey,
			GroupPath:    "storage/oss",
			Label:        "OSS Endpoint",
			Description:  "阿里云 OSS Endpoint，例如 https://oss-cn-beijing.aliyuncs.com。",
			DefaultValue: strings.TrimSpace(s.defaultStorage.OSS.Endpoint),
			VisibleWhen:  ossVisibleWhen,
			Sort:         20,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSBucketKey,
			GroupPath:    "storage/oss",
			Label:        "Bucket",
			Description:  "用于保存资源的 OSS Bucket 名称。",
			DefaultValue: strings.TrimSpace(s.defaultStorage.OSS.Bucket),
			VisibleWhen:  ossVisibleWhen,
			Sort:         30,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSPrefixKey,
			GroupPath:    "storage/oss",
			Label:        "对象前缀",
			Description:  "写入 OSS 时附加的对象前缀，可留空。",
			DefaultValue: strings.Trim(strings.TrimSpace(s.defaultStorage.OSS.Prefix), "/"),
			VisibleWhen:  ossVisibleWhen,
			Sort:         40,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSPublicBaseURLKey,
			GroupPath:    "storage/oss",
			Label:        "公开访问域名",
			Description:  "前台和后台返回给用户的资源访问基地址。",
			DefaultValue: strings.TrimRight(strings.TrimSpace(s.defaultStorage.OSS.PublicBaseURL), "/"),
			VisibleWhen:  ossVisibleWhen,
			Sort:         50,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSAccessKeyIDKey,
			GroupPath:    "storage/oss",
			Label:        "Access Key ID",
			Description:  "阿里云 RAM AccessKey ID。",
			DefaultValue: strings.TrimSpace(s.defaultStorage.OSS.AccessKeyID),
			VisibleWhen:  ossVisibleWhen,
			Sort:         60,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSAccessKeySecretKey,
			GroupPath:    "storage/oss",
			Label:        "Access Key Secret",
			Description:  "阿里云 RAM AccessKey Secret，留空表示保持原值。",
			DefaultValue: strings.TrimSpace(s.defaultStorage.OSS.AccessKeySecret),
			IsSensitive:  true,
			Meta:         passwordMeta,
			VisibleWhen:  ossVisibleWhen,
			Sort:         70,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          storageOSSSecurityTokenKey,
			GroupPath:    "storage/oss",
			Label:        "Security Token",
			Description:  "使用临时凭证时填写 STS Security Token，可留空。",
			DefaultValue: strings.TrimSpace(s.defaultStorage.OSS.SecurityToken),
			IsSensitive:  true,
			Meta:         textareaMeta,
			VisibleWhen:  ossVisibleWhen,
			Sort:         80,
		}),
	}
}

func (s *Service) mergeBuiltinConfigs(items []domainconfig.SysConfig, keys []string) []domainconfig.SysConfig {
	builtins := s.filteredBuiltinConfigs(keys)
	if len(builtins) == 0 {
		return items
	}
	merged := make([]domainconfig.SysConfig, 0, len(items)+len(builtins))
	seen := make(map[string]struct{}, len(items))
	builtinMap := make(map[string]domainconfig.SysConfig, len(builtins))
	for _, builtin := range builtins {
		builtinMap[builtin.Key] = builtin
	}
	for _, item := range items {
		if builtin, ok := builtinMap[item.Key]; ok {
			merged = append(merged, applyBuiltinSchema(item, builtin))
		} else {
			merged = append(merged, item)
		}
		seen[item.Key] = struct{}{}
	}
	for _, builtin := range builtins {
		if _, ok := seen[builtin.Key]; ok {
			continue
		}
		merged = append(merged, builtin)
	}
	return merged
}

func (s *Service) builtinConfigByKey(key string) (domainconfig.SysConfig, bool) {
	for _, item := range s.filteredBuiltinConfigs([]string{key}) {
		if item.Key == key {
			return item, true
		}
	}
	return domainconfig.SysConfig{}, false
}

func (s *Service) filteredBuiltinConfigs(keys []string) []domainconfig.SysConfig {
	all := append(s.builtinUploadConfigs(), s.builtinStorageBackendConfigs()...)
	all = append(all, s.builtinMediaConfigs()...)
	if len(keys) == 0 {
		return all
	}
	allow := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			allow[key] = struct{}{}
		}
	}
	filtered := make([]domainconfig.SysConfig, 0, len(all))
	for _, item := range all {
		if _, ok := allow[item.Key]; ok {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (s *Service) builtinMediaConfigs() []domainconfig.SysConfig {
	passwordMeta := mustRawJSON(`{"inputType":"password"}`)
	timeout := s.defaultMedia.TMDBTimeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return []domainconfig.SysConfig{
		newBuiltinConfig(builtinConfigSpec{
			Key:          mediaTMDBAPIKeyKey,
			GroupPath:    "media/tmdb",
			Label:        "TMDB API Key",
			Description:  "用于搜索电影和电视剧信息。后台配置优先于环境变量 TMDB_API_KEY。",
			DefaultValue: strings.TrimSpace(s.defaultMedia.TMDBAPIKey),
			IsSensitive:  true,
			Meta:         passwordMeta,
			Sort:         10,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          mediaTMDBLanguageKey,
			GroupPath:    "media/tmdb",
			Label:        "TMDB 语言",
			Description:  "TMDB 搜索结果语言，例如 zh-CN 或 en-US。",
			DefaultValue: strings.TrimSpace(s.defaultMedia.TMDBLanguage),
			Sort:         20,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          mediaTMDBBaseURLKey,
			GroupPath:    "media/tmdb",
			Label:        "API 地址",
			Description:  "TMDB API 基地址。服务器无法直连 TMDB 时，可填写兼容的反向代理地址。",
			DefaultValue: strings.TrimRight(strings.TrimSpace(s.defaultMedia.TMDBBaseURL), "/"),
			Sort:         30,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          mediaTMDBImageURLKey,
			GroupPath:    "media/tmdb",
			Label:        "图片地址",
			Description:  "海报与背景图的访问基地址，可配置为图片代理地址。",
			DefaultValue: strings.TrimRight(strings.TrimSpace(s.defaultMedia.TMDBImageBaseURL), "/"),
			Sort:         40,
		}),
		newBuiltinConfig(builtinConfigSpec{
			Key:          mediaTMDBTimeoutKey,
			GroupPath:    "media/tmdb",
			Label:        "请求超时（秒）",
			Description:  "TMDB 搜索请求的最长等待时间，范围 1–120 秒。",
			ValueType:    valueTypeNumber,
			DefaultValue: strconv.Itoa(int(timeout.Seconds())),
			Sort:         50,
		}),
	}
}

type builtinConfigSpec struct {
	Key          string
	GroupPath    string
	Label        string
	Description  string
	ValueType    string
	DefaultValue string
	IsSensitive  bool
	EnumOptions  json.RawMessage
	VisibleWhen  json.RawMessage
	Meta         json.RawMessage
	Sort         int
}

func newStringBuiltinConfig(key string, groupPath string, label string, description string, defaultValue string, sort int) domainconfig.SysConfig {
	valueType := valueTypeString
	if key == uploadMaxSizeKey {
		valueType = valueTypeNumber
	}
	return newBuiltinConfig(builtinConfigSpec{
		Key:          key,
		GroupPath:    groupPath,
		Label:        label,
		Description:  description,
		ValueType:    valueType,
		DefaultValue: defaultValue,
		Sort:         sort,
	})
}

func newBuiltinConfig(spec builtinConfigSpec) domainconfig.SysConfig {
	defaultValue := strings.TrimSpace(spec.DefaultValue)
	cfg := domainconfig.SysConfig{
		Key:         spec.Key,
		GroupPath:   spec.GroupPath,
		Label:       spec.Label,
		Description: spec.Description,
		ValueType:   normalizeValueType(spec.ValueType),
		IsSensitive: spec.IsSensitive,
		EnumOptions: emptyJSONArray,
		VisibleWhen: emptyJSONArray,
		Meta:        emptyJSONObject,
		Sort:        spec.Sort,
	}
	if !cfg.IsSensitive && (defaultValue != "" || cfg.ValueType == valueTypeString || cfg.ValueType == valueTypeEnum) {
		cfg.DefaultValue = &defaultValue
	}
	if len(spec.EnumOptions) > 0 {
		cfg.EnumOptions = spec.EnumOptions
	}
	if len(spec.VisibleWhen) > 0 {
		cfg.VisibleWhen = spec.VisibleWhen
	}
	if len(spec.Meta) > 0 {
		cfg.Meta = spec.Meta
	}
	return cfg
}

func normalizeUploadManagedDir(value string, fallback string) string {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return strings.Trim(strings.TrimSpace(fallback), "/")
	}
	return trimmed
}

func normalizeStorageProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "local":
		return "local"
	case "aliyun_oss":
		return "aliyun_oss"
	default:
		return "local"
	}
}

func mustRawJSON(value string) json.RawMessage {
	return json.RawMessage(strings.TrimSpace(value))
}

func applyBuiltinSchema(current domainconfig.SysConfig, builtin domainconfig.SysConfig) domainconfig.SysConfig {
	current.GroupPath = builtin.GroupPath
	current.Label = builtin.Label
	current.Description = builtin.Description
	current.ValueType = builtin.ValueType
	current.EnumOptions = builtin.EnumOptions
	current.DefaultValue = builtin.DefaultValue
	current.VisibleWhen = builtin.VisibleWhen
	current.Sort = builtin.Sort
	current.Meta = builtin.Meta
	current.IsSensitive = builtin.IsSensitive
	return current
}
