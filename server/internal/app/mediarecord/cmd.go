package mediarecord

import "time"

type CreateCmd struct {
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
}

type UpdateCmd struct {
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
}
