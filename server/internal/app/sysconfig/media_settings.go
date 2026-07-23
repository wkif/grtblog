package sysconfig

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/config"
)

const (
	mediaTMDBAPIKeyKey   = "media.tmdb.apiKey"
	mediaTMDBLanguageKey = "media.tmdb.language"
	mediaTMDBBaseURLKey  = "media.tmdb.baseURL"
	mediaTMDBImageURLKey = "media.tmdb.imageBaseURL"
	mediaTMDBTimeoutKey  = "media.tmdb.timeoutSeconds"
)

// MediaSettings returns the current media integration settings. Values saved in
// sys_config take precedence over environment-backed defaults.
func (s *Service) MediaSettings(ctx context.Context) config.MediaConfig {
	settings := config.MediaConfig{
		TMDBAPIKey:       strings.TrimSpace(s.defaultMedia.TMDBAPIKey),
		TMDBLanguage:     strings.TrimSpace(s.defaultMedia.TMDBLanguage),
		TMDBBaseURL:      strings.TrimRight(strings.TrimSpace(s.defaultMedia.TMDBBaseURL), "/"),
		TMDBImageBaseURL: strings.TrimRight(strings.TrimSpace(s.defaultMedia.TMDBImageBaseURL), "/"),
		TMDBTimeout:      s.defaultMedia.TMDBTimeout,
	}
	if settings.TMDBLanguage == "" {
		settings.TMDBLanguage = "zh-CN"
	}
	if settings.TMDBBaseURL == "" {
		settings.TMDBBaseURL = "https://api.themoviedb.org/3"
	}
	if settings.TMDBImageBaseURL == "" {
		settings.TMDBImageBaseURL = "https://image.tmdb.org/t/p/w780"
	}
	if settings.TMDBTimeout <= 0 {
		settings.TMDBTimeout = 20 * time.Second
	}

	items, err := s.repo.List(ctx, []string{
		mediaTMDBAPIKeyKey,
		mediaTMDBLanguageKey,
		mediaTMDBBaseURLKey,
		mediaTMDBImageURLKey,
		mediaTMDBTimeoutKey,
	})
	if err != nil {
		return settings
	}
	for _, item := range items {
		value := strings.TrimSpace(item.Value)
		if value == "" {
			continue
		}
		switch item.Key {
		case mediaTMDBAPIKeyKey:
			settings.TMDBAPIKey = value
		case mediaTMDBLanguageKey:
			settings.TMDBLanguage = value
		case mediaTMDBBaseURLKey:
			settings.TMDBBaseURL = strings.TrimRight(value, "/")
		case mediaTMDBImageURLKey:
			settings.TMDBImageBaseURL = strings.TrimRight(value, "/")
		case mediaTMDBTimeoutKey:
			seconds, parseErr := strconv.Atoi(value)
			if parseErr == nil && seconds > 0 && seconds <= 120 {
				settings.TMDBTimeout = time.Duration(seconds) * time.Second
			}
		}
	}
	return settings
}
