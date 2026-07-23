package mediarecord

import "time"

const (
	TypeMovie = "movie"
	TypeTV    = "tv"

	StatusPlanned   = "planned"
	StatusWatching  = "watching"
	StatusCompleted = "completed"
	StatusDropped   = "dropped"
)

type Record struct {
	ID             int64
	Title          string
	OriginalTitle  *string
	MediaType      string
	Provider       string
	ProviderID     *string
	Poster         *string
	Backdrop       *string
	Overview       *string
	ReleaseDate    *time.Time
	RuntimeMinutes *int
	TotalEpisodes  *int
	Status         string
	Progress       int
	ProgressTotal  *int
	Rating         *float64
	Note           *string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	IsPublished    bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SearchResult struct {
	ProviderID     string
	Title          string
	OriginalTitle  string
	MediaType      string
	Poster         string
	Backdrop       string
	Overview       string
	ReleaseDate    string
	RuntimeMinutes int
	TotalEpisodes  int
}
