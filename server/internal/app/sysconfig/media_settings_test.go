package sysconfig

import (
	"context"
	"testing"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

func TestMediaSettingsPrefersSysConfigValues(t *testing.T) {
	t.Parallel()

	repo := &fakeSysConfigRepo{items: map[string]domainconfig.SysConfig{
		mediaTMDBAPIKeyKey:   {Key: mediaTMDBAPIKeyKey, Value: "db-key"},
		mediaTMDBLanguageKey: {Key: mediaTMDBLanguageKey, Value: "en-US"},
		mediaTMDBBaseURLKey:  {Key: mediaTMDBBaseURLKey, Value: "https://tmdb.example.test/3/"},
		mediaTMDBImageURLKey: {Key: mediaTMDBImageURLKey, Value: "https://images.example.test/w780/"},
		mediaTMDBTimeoutKey:  {Key: mediaTMDBTimeoutKey, Value: "45"},
	}}
	svc := NewServiceWithMedia(repo, config.TurnstileConfig{}, config.StorageConfig{}, config.MediaConfig{
		TMDBAPIKey:   "env-key",
		TMDBLanguage: "zh-CN",
	}, appEvent.NopBus{})

	settings := svc.MediaSettings(context.Background())
	if settings.TMDBAPIKey != "db-key" {
		t.Fatalf("settings.TMDBAPIKey = %q, want db-key", settings.TMDBAPIKey)
	}
	if settings.TMDBLanguage != "en-US" {
		t.Fatalf("settings.TMDBLanguage = %q, want en-US", settings.TMDBLanguage)
	}
	if settings.TMDBBaseURL != "https://tmdb.example.test/3" {
		t.Fatalf("settings.TMDBBaseURL = %q", settings.TMDBBaseURL)
	}
	if settings.TMDBImageBaseURL != "https://images.example.test/w780" {
		t.Fatalf("settings.TMDBImageBaseURL = %q", settings.TMDBImageBaseURL)
	}
	if settings.TMDBTimeout != 45*time.Second {
		t.Fatalf("settings.TMDBTimeout = %s, want 45s", settings.TMDBTimeout)
	}
}

func TestMediaSettingsFallsBackToDefaults(t *testing.T) {
	t.Parallel()

	svc := NewServiceWithMedia(&fakeSysConfigRepo{}, config.TurnstileConfig{}, config.StorageConfig{}, config.MediaConfig{
		TMDBAPIKey: "env-key",
	}, appEvent.NopBus{})

	settings := svc.MediaSettings(context.Background())
	if settings.TMDBAPIKey != "env-key" {
		t.Fatalf("settings.TMDBAPIKey = %q, want env-key", settings.TMDBAPIKey)
	}
	if settings.TMDBLanguage != "zh-CN" {
		t.Fatalf("settings.TMDBLanguage = %q, want zh-CN", settings.TMDBLanguage)
	}
}

func TestListConfigsIncludesBuiltinMediaSettings(t *testing.T) {
	t.Parallel()

	svc := NewServiceWithMedia(&fakeSysConfigRepo{}, config.TurnstileConfig{}, config.StorageConfig{}, config.MediaConfig{
		TMDBAPIKey:   "env-key",
		TMDBLanguage: "zh-CN",
	}, appEvent.NopBus{})
	items, err := svc.ListConfigs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListConfigs() error = %v", err)
	}

	var apiKey domainconfig.SysConfig
	for _, item := range items {
		if item.Key == mediaTMDBAPIKeyKey {
			apiKey = item
			break
		}
	}
	if apiKey.Key == "" {
		t.Fatalf("ListConfigs() missing builtin key %q", mediaTMDBAPIKeyKey)
	}
	if !apiKey.IsSensitive {
		t.Fatal("TMDB API key should be sensitive")
	}
	if apiKey.DefaultValue != nil {
		t.Fatal("sensitive TMDB API key should not expose a default value")
	}
}
