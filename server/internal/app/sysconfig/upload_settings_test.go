package sysconfig

import (
	"context"
	"encoding/json"
	"testing"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

func TestListConfigsIncludesBuiltinUploadSettings(t *testing.T) {
	t.Parallel()

	svc := NewService(&fakeSysConfigRepo{}, config.TurnstileConfig{}, config.StorageConfig{}, appEvent.NopBus{})
	items, err := svc.ListConfigs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListConfigs() error = %v", err)
	}
	if !hasSysConfigKey(items, uploadCacheDirKey) {
		t.Fatalf("ListConfigs() missing builtin key %q", uploadCacheDirKey)
	}
	if !hasSysConfigKey(items, uploadImageDirKey) {
		t.Fatalf("ListConfigs() missing builtin key %q", uploadImageDirKey)
	}
	if !hasSysConfigKey(items, storageProviderKey) {
		t.Fatalf("ListConfigs() missing builtin key %q", storageProviderKey)
	}
	if !hasSysConfigKey(items, storageOSSEndpointKey) {
		t.Fatalf("ListConfigs() missing builtin key %q", storageOSSEndpointKey)
	}
}

func TestUpdateConfigsCreatesBuiltinUploadMetadata(t *testing.T) {
	t.Parallel()

	repo := &fakeSysConfigRepo{}
	svc := NewService(repo, config.TurnstileConfig{}, config.StorageConfig{}, appEvent.NopBus{})
	valueType := "string"
	groupPath := "storage/upload"
	label := "草稿缓存目录"
	value := json.RawMessage(`"custom/cache"`)
	items, err := svc.UpdateConfigs(context.Background(), []UpdateItem{
		{
			Key:   uploadCacheDirKey,
			Value: &value,
		},
	})
	if err != nil {
		t.Fatalf("UpdateConfigs() error = %v", err)
	}
	if !hasSysConfigKey(items, uploadCacheDirKey) {
		t.Fatalf("UpdateConfigs() missing updated key %q", uploadCacheDirKey)
	}
	saved, err := repo.GetByKey(context.Background(), uploadCacheDirKey)
	if err != nil {
		t.Fatalf("GetByKey() error = %v", err)
	}
	if saved.GroupPath != groupPath {
		t.Fatalf("saved.GroupPath = %q, want %q", saved.GroupPath, groupPath)
	}
	if saved.Label != label {
		t.Fatalf("saved.Label = %q, want %q", saved.Label, label)
	}
	if saved.ValueType != valueType {
		t.Fatalf("saved.ValueType = %q, want %q", saved.ValueType, valueType)
	}
}

func TestStorageSettingsPrefersSysConfigValues(t *testing.T) {
	t.Parallel()

	repo := &fakeSysConfigRepo{
		items: map[string]domainconfig.SysConfig{
			storageProviderKey: {
				Key:   storageProviderKey,
				Value: "aliyun_oss",
			},
			storageOSSEndpointKey: {
				Key:   storageOSSEndpointKey,
				Value: "https://oss-cn-shanghai.aliyuncs.com",
			},
			storageOSSBucketKey: {
				Key:   storageOSSBucketKey,
				Value: "blog-bucket",
			},
			storageOSSAccessKeySecretKey: {
				Key:   storageOSSAccessKeySecretKey,
				Value: "secret-from-db",
			},
			uploadCacheDirKey: {
				Key:   uploadCacheDirKey,
				Value: "custom/cache",
			},
		},
	}
	svc := NewService(repo, config.TurnstileConfig{}, config.StorageConfig{
		Provider: "local",
		CacheDir: "blog/cache",
		OSS: config.OSSConfig{
			Endpoint:        "https://oss-cn-beijing.aliyuncs.com",
			Bucket:          "fallback-bucket",
			AccessKeySecret: "secret-from-env",
		},
	}, appEvent.NopBus{})

	settings := svc.StorageSettings(context.Background())
	if settings.Provider != "aliyun_oss" {
		t.Fatalf("settings.Provider = %q, want aliyun_oss", settings.Provider)
	}
	if settings.OSS.Endpoint != "https://oss-cn-shanghai.aliyuncs.com" {
		t.Fatalf("settings.OSS.Endpoint = %q, want db endpoint", settings.OSS.Endpoint)
	}
	if settings.OSS.Bucket != "blog-bucket" {
		t.Fatalf("settings.OSS.Bucket = %q, want blog-bucket", settings.OSS.Bucket)
	}
	if settings.OSS.AccessKeySecret != "secret-from-db" {
		t.Fatalf("settings.OSS.AccessKeySecret = %q, want secret-from-db", settings.OSS.AccessKeySecret)
	}
	if settings.CacheDir != "custom/cache" {
		t.Fatalf("settings.CacheDir = %q, want custom/cache", settings.CacheDir)
	}
}

func TestStorageSettingsWithOverridesUsesDraftValues(t *testing.T) {
	t.Parallel()

	repo := &fakeSysConfigRepo{
		items: map[string]domainconfig.SysConfig{
			storageProviderKey: {
				Key:   storageProviderKey,
				Value: "local",
			},
			storageOSSEndpointKey: {
				Key:   storageOSSEndpointKey,
				Value: "https://oss-cn-beijing.aliyuncs.com",
			},
		},
	}
	svc := NewService(repo, config.TurnstileConfig{}, config.StorageConfig{}, appEvent.NopBus{})

	provider := json.RawMessage(`"aliyun_oss"`)
	endpoint := json.RawMessage(`"https://oss-cn-hangzhou.aliyuncs.com"`)
	secret := json.RawMessage(`"draft-secret"`)
	settings, err := svc.StorageSettingsWithOverrides(context.Background(), []UpdateItem{
		{Key: storageProviderKey, Value: &provider},
		{Key: storageOSSEndpointKey, Value: &endpoint},
		{Key: storageOSSAccessKeySecretKey, Value: &secret},
	})
	if err != nil {
		t.Fatalf("StorageSettingsWithOverrides() error = %v", err)
	}
	if settings.Provider != "aliyun_oss" {
		t.Fatalf("settings.Provider = %q, want aliyun_oss", settings.Provider)
	}
	if settings.OSS.Endpoint != "https://oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("settings.OSS.Endpoint = %q, want override endpoint", settings.OSS.Endpoint)
	}
	if settings.OSS.AccessKeySecret != "draft-secret" {
		t.Fatalf("settings.OSS.AccessKeySecret = %q, want draft-secret", settings.OSS.AccessKeySecret)
	}
}

func TestListConfigsOverlaysBuiltinSchemaOnLegacyStorageProvider(t *testing.T) {
	t.Parallel()

	repo := &fakeSysConfigRepo{
		items: map[string]domainconfig.SysConfig{
			storageProviderKey: {
				Key:       storageProviderKey,
				Value:     "aliyun_oss",
				ValueType: valueTypeString,
			},
		},
	}
	svc := NewService(repo, config.TurnstileConfig{}, config.StorageConfig{}, appEvent.NopBus{})

	items, err := svc.ListConfigs(context.Background(), []string{storageProviderKey})
	if err != nil {
		t.Fatalf("ListConfigs() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ValueType != valueTypeEnum {
		t.Fatalf("items[0].ValueType = %q, want %q", items[0].ValueType, valueTypeEnum)
	}
}

func TestUpdateConfigsRepairsLegacyStorageProviderValueType(t *testing.T) {
	t.Parallel()

	repo := &fakeSysConfigRepo{
		items: map[string]domainconfig.SysConfig{
			storageProviderKey: {
				Key:       storageProviderKey,
				Value:     "local",
				ValueType: valueTypeString,
			},
		},
	}
	svc := NewService(repo, config.TurnstileConfig{}, config.StorageConfig{}, appEvent.NopBus{})
	provider := json.RawMessage(`"aliyun_oss"`)

	_, err := svc.UpdateConfigs(context.Background(), []UpdateItem{
		{
			Key:   storageProviderKey,
			Value: &provider,
		},
	})
	if err != nil {
		t.Fatalf("UpdateConfigs() error = %v", err)
	}
	saved, err := repo.GetByKey(context.Background(), storageProviderKey)
	if err != nil {
		t.Fatalf("GetByKey() error = %v", err)
	}
	if saved.ValueType != valueTypeEnum {
		t.Fatalf("saved.ValueType = %q, want %q", saved.ValueType, valueTypeEnum)
	}
	if saved.Value != "aliyun_oss" {
		t.Fatalf("saved.Value = %q, want aliyun_oss", saved.Value)
	}
}

func hasSysConfigKey(items []domainconfig.SysConfig, key string) bool {
	for _, item := range items {
		if item.Key == key {
			return true
		}
	}
	return false
}

type fakeSysConfigRepo struct {
	items map[string]domainconfig.SysConfig
}

func (r *fakeSysConfigRepo) GetByKey(_ context.Context, key string) (*domainconfig.SysConfig, error) {
	if r.items == nil {
		return nil, domainconfig.ErrSysConfigNotFound
	}
	item, ok := r.items[key]
	if !ok {
		return nil, domainconfig.ErrSysConfigNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (r *fakeSysConfigRepo) List(_ context.Context, keys []string) ([]domainconfig.SysConfig, error) {
	if r.items == nil {
		return nil, nil
	}
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

func (r *fakeSysConfigRepo) Upsert(_ context.Context, configs []domainconfig.SysConfig) error {
	if r.items == nil {
		r.items = make(map[string]domainconfig.SysConfig, len(configs))
	}
	for _, cfg := range configs {
		r.items[cfg.Key] = cfg
	}
	return nil
}
