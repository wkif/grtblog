package mediarecord

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTMDBClientUsesDynamicSettings(t *testing.T) {
	t.Parallel()

	client := NewDynamicTMDBClient(func(context.Context) TMDBSettings {
		return TMDBSettings{
			APIKey:       "test-key",
			Language:     "en-US",
			BaseURL:      "https://tmdb.example.test/3",
			ImageBaseURL: "https://images.example.test/w780",
			Timeout:      time.Second,
		}
	})
	client.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://tmdb.example.test/3/search/multi?api_key=test-key&include_adult=false&language=en-US&query=Example" {
			t.Errorf("request URL = %q", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"results":[{"id":42,"media_type":"movie","title":"Example","poster_path":"/poster.jpg"}]}`)),
		}, nil
	})

	results, err := client.Search(context.Background(), "Example", "movie")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Poster != "https://images.example.test/w780/poster.jpg" {
		t.Fatalf("results[0].Poster = %q", results[0].Poster)
	}
}

func TestTMDBClientTimeoutDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()

	const apiKey = "secret-api-key"
	client := NewDynamicTMDBClient(func(context.Context) TMDBSettings {
		return TMDBSettings{APIKey: apiKey, BaseURL: "https://tmdb.example.test/3", Timeout: time.Millisecond}
	})
	client.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})

	_, err := client.Search(context.Background(), "Example", "")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Search() error = %v, want context deadline exceeded", err)
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("Search() error leaked API key: %v", err)
	}
}

func TestTMDBClientDetailsLoadsRuntimeAndEpisodeCount(t *testing.T) {
	t.Parallel()

	client := NewDynamicTMDBClient(func(context.Context) TMDBSettings {
		return TMDBSettings{
			APIKey:       "test-key",
			Language:     "zh-CN",
			BaseURL:      "https://tmdb.example.test/3",
			ImageBaseURL: "https://images.example.test/w780",
			Timeout:      time.Second,
		}
	})
	client.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"id":42,"title":"Example Movie","runtime":126}`
		if req.URL.Path == "/3/tv/84" {
			body = `{"id":84,"name":"Example Series","number_of_episodes":24}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	movie, err := client.Details(context.Background(), "42", "movie")
	if err != nil {
		t.Fatalf("movie Details() error = %v", err)
	}
	if movie.RuntimeMinutes != 126 {
		t.Fatalf("movie.RuntimeMinutes = %d, want 126", movie.RuntimeMinutes)
	}

	series, err := client.Details(context.Background(), "84", "tv")
	if err != nil {
		t.Fatalf("tv Details() error = %v", err)
	}
	if series.TotalEpisodes != 24 {
		t.Fatalf("series.TotalEpisodes = %d, want 24", series.TotalEpisodes)
	}
}
