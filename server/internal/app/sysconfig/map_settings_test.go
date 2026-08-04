package sysconfig

import (
	"context"
	"testing"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

func TestMapSettingsDefaultsToOSM(t *testing.T) {
	t.Parallel()

	svc := NewService(&fakeSysConfigRepo{}, config.TurnstileConfig{}, config.StorageConfig{}, appEvent.NopBus{})
	settings, err := svc.MapSettings(context.Background())
	if err != nil {
		t.Fatalf("MapSettings() error = %v", err)
	}
	if settings.Provider != "osm" || settings.TiandituLayer != "vector" || settings.TiandituKey != "" {
		t.Fatalf("MapSettings() = %+v, want OSM defaults", settings)
	}
}

func TestMapSettingsLoadsTiandituConfig(t *testing.T) {
	t.Parallel()

	repo := &fakeSysConfigRepo{items: map[string]domainconfig.SysConfig{
		"map.provider":       {Key: "map.provider", Value: "tianditu"},
		"map.tianditu.key":   {Key: "map.tianditu.key", Value: "test-key"},
		"map.tianditu.layer": {Key: "map.tianditu.layer", Value: "imagery"},
	}}
	svc := NewService(repo, config.TurnstileConfig{}, config.StorageConfig{}, appEvent.NopBus{})
	settings, err := svc.MapSettings(context.Background())
	if err != nil {
		t.Fatalf("MapSettings() error = %v", err)
	}
	if settings.Provider != "tianditu" || settings.TiandituLayer != "imagery" || settings.TiandituKey != "test-key" {
		t.Fatalf("MapSettings() = %+v, want configured Tianditu values", settings)
	}
}
