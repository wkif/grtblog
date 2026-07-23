package mediarecord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/mediarecord"
)

type TMDBClient struct {
	settings func(context.Context) TMDBSettings
	client   *http.Client
}

type TMDBSettings struct {
	APIKey       string
	Language     string
	BaseURL      string
	ImageBaseURL string
	Timeout      time.Duration
}

func NewTMDBClient(apiKey, language string) *TMDBClient {
	return NewDynamicTMDBClient(func(context.Context) TMDBSettings {
		return TMDBSettings{APIKey: apiKey, Language: language}
	})
}

func NewDynamicTMDBClient(settings func(context.Context) TMDBSettings) *TMDBClient {
	return &TMDBClient{settings: settings, client: &http.Client{}}
}

type tmdbSearchResponse struct {
	Results []tmdbSearchItem `json:"results"`
}
type tmdbSearchItem struct {
	ID               int64  `json:"id"`
	MediaType        string `json:"media_type"`
	Title            string `json:"title"`
	Name             string `json:"name"`
	OriginalTitle    string `json:"original_title"`
	OriginalName     string `json:"original_name"`
	PosterPath       string `json:"poster_path"`
	BackdropPath     string `json:"backdrop_path"`
	Overview         string `json:"overview"`
	ReleaseDate      string `json:"release_date"`
	FirstAirDate     string `json:"first_air_date"`
	Runtime          int    `json:"runtime"`
	NumberOfEpisodes int    `json:"number_of_episodes"`
}

type tmdbDetailResponse struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Name             string `json:"name"`
	OriginalTitle    string `json:"original_title"`
	OriginalName     string `json:"original_name"`
	PosterPath       string `json:"poster_path"`
	BackdropPath     string `json:"backdrop_path"`
	Overview         string `json:"overview"`
	ReleaseDate      string `json:"release_date"`
	FirstAirDate     string `json:"first_air_date"`
	Runtime          int    `json:"runtime"`
	NumberOfEpisodes int    `json:"number_of_episodes"`
}

func (c *TMDBClient) Search(ctx context.Context, query, mediaType string) ([]domain.SearchResult, error) {
	settings := c.resolveSettings(ctx)
	if settings.APIKey == "" {
		return []domain.SearchResult{}, nil
	}
	q := make(url.Values)
	q.Set("api_key", settings.APIKey)
	q.Set("query", query)
	q.Set("language", settings.Language)
	q.Set("include_adult", "false")
	var payload tmdbSearchResponse
	if err := c.getJSON(ctx, settings, "/search/multi", q, &payload); err != nil {
		return nil, err
	}
	result := make([]domain.SearchResult, 0, len(payload.Results))
	for _, item := range payload.Results {
		kind := item.MediaType
		if kind != domain.TypeMovie && kind != domain.TypeTV {
			continue
		}
		if mediaType != "" && mediaType != kind {
			continue
		}
		title, original, date := item.Title, item.OriginalTitle, item.ReleaseDate
		if kind == domain.TypeTV {
			title, original, date = item.Name, item.OriginalName, item.FirstAirDate
		}
		result = append(result, domain.SearchResult{ProviderID: strconv.FormatInt(item.ID, 10), Title: title, OriginalTitle: original, MediaType: kind, Poster: imageURL(settings.ImageBaseURL, item.PosterPath), Backdrop: imageURL(settings.ImageBaseURL, item.BackdropPath), Overview: item.Overview, ReleaseDate: date, RuntimeMinutes: item.Runtime, TotalEpisodes: item.NumberOfEpisodes})
	}
	return result, nil
}

func (c *TMDBClient) Details(ctx context.Context, providerID string, mediaType string) (domain.SearchResult, error) {
	settings := c.resolveSettings(ctx)
	if settings.APIKey == "" {
		return domain.SearchResult{}, fmt.Errorf("tmdb api key is missing")
	}
	q := make(url.Values)
	q.Set("api_key", settings.APIKey)
	q.Set("language", settings.Language)
	var payload tmdbDetailResponse
	if err := c.getJSON(ctx, settings, "/"+mediaType+"/"+url.PathEscape(providerID), q, &payload); err != nil {
		return domain.SearchResult{}, err
	}
	title, original, date := payload.Title, payload.OriginalTitle, payload.ReleaseDate
	if mediaType == domain.TypeTV {
		title, original, date = payload.Name, payload.OriginalName, payload.FirstAirDate
	}
	return domain.SearchResult{
		ProviderID:     strconv.FormatInt(payload.ID, 10),
		Title:          title,
		OriginalTitle:  original,
		MediaType:      mediaType,
		Poster:         imageURL(settings.ImageBaseURL, payload.PosterPath),
		Backdrop:       imageURL(settings.ImageBaseURL, payload.BackdropPath),
		Overview:       payload.Overview,
		ReleaseDate:    date,
		RuntimeMinutes: payload.Runtime,
		TotalEpisodes:  payload.NumberOfEpisodes,
	}, nil
}

func (c *TMDBClient) resolveSettings(ctx context.Context) TMDBSettings {
	settings := TMDBSettings{
		Language:     "zh-CN",
		BaseURL:      "https://api.themoviedb.org/3",
		ImageBaseURL: "https://image.tmdb.org/t/p/w780",
		Timeout:      20 * time.Second,
	}
	if c.settings != nil {
		settings = c.settings(ctx)
	}
	settings.APIKey = strings.TrimSpace(settings.APIKey)
	settings.Language = strings.TrimSpace(settings.Language)
	settings.BaseURL = strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")
	settings.ImageBaseURL = strings.TrimRight(strings.TrimSpace(settings.ImageBaseURL), "/")
	if settings.Language == "" {
		settings.Language = "zh-CN"
	}
	if settings.BaseURL == "" {
		settings.BaseURL = "https://api.themoviedb.org/3"
	}
	if settings.ImageBaseURL == "" {
		settings.ImageBaseURL = "https://image.tmdb.org/t/p/w780"
	}
	if settings.Timeout <= 0 {
		settings.Timeout = 20 * time.Second
	}
	return settings
}

func (c *TMDBClient) getJSON(ctx context.Context, settings TMDBSettings, path string, query url.Values, target any) error {
	u, err := url.Parse(settings.BaseURL + path)
	if err != nil {
		return err
	}
	u.RawQuery = query.Encode()
	requestContext, cancel := context.WithTimeout(ctx, settings.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestContext, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if errors.Is(urlErr.Err, context.DeadlineExceeded) {
				return fmt.Errorf("tmdb request timed out after %s: %w", settings.Timeout, context.DeadlineExceeded)
			}
			var netErr net.Error
			if errors.As(urlErr.Err, &netErr) && netErr.Timeout() {
				return fmt.Errorf("tmdb request timed out after %s: %w", settings.Timeout, context.DeadlineExceeded)
			}
			return fmt.Errorf("tmdb request failed: %w", urlErr.Err)
		}
		return fmt.Errorf("tmdb request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("tmdb request returned %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func imageURL(baseURL, path string) string {
	if path == "" {
		return ""
	}
	return baseURL + path
}
