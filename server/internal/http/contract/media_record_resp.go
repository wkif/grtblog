package contract

import (
	"time"

	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/mediarecord"
)

type MediaRecordResp struct {
	ID             int64      `json:"id"`
	Title          string     `json:"title"`
	OriginalTitle  *string    `json:"originalTitle,omitempty"`
	MediaType      string     `json:"mediaType"`
	Provider       string     `json:"provider"`
	ProviderID     *string    `json:"providerId,omitempty"`
	Poster         *string    `json:"poster,omitempty"`
	Backdrop       *string    `json:"backdrop,omitempty"`
	Overview       *string    `json:"overview,omitempty"`
	ReleaseDate    *time.Time `json:"releaseDate,omitempty"`
	RuntimeMinutes *int       `json:"runtimeMinutes,omitempty"`
	TotalEpisodes  *int       `json:"totalEpisodes,omitempty"`
	Status         string     `json:"status"`
	Progress       int        `json:"progress"`
	ProgressTotal  *int       `json:"progressTotal,omitempty"`
	Rating         *float64   `json:"rating,omitempty"`
	Note           *string    `json:"note,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	IsPublished    bool       `json:"isPublished"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type MediaRecordListResp struct {
	Items []MediaRecordResp `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

type MediaSearchResultResp struct {
	ProviderID     string `json:"providerId"`
	Title          string `json:"title"`
	OriginalTitle  string `json:"originalTitle"`
	MediaType      string `json:"mediaType"`
	Poster         string `json:"poster,omitempty"`
	Backdrop       string `json:"backdrop,omitempty"`
	Overview       string `json:"overview,omitempty"`
	ReleaseDate    string `json:"releaseDate,omitempty"`
	RuntimeMinutes int    `json:"runtimeMinutes,omitempty"`
	TotalEpisodes  int    `json:"totalEpisodes,omitempty"`
}

func MediaRecordResponse(record *domain.Record) MediaRecordResp {
	return MediaRecordResp{ID: record.ID, Title: record.Title, OriginalTitle: record.OriginalTitle, MediaType: record.MediaType, Provider: record.Provider, ProviderID: record.ProviderID, Poster: record.Poster, Backdrop: record.Backdrop, Overview: record.Overview, ReleaseDate: record.ReleaseDate, RuntimeMinutes: record.RuntimeMinutes, TotalEpisodes: record.TotalEpisodes, Status: record.Status, Progress: record.Progress, ProgressTotal: record.ProgressTotal, Rating: record.Rating, Note: record.Note, StartedAt: record.StartedAt, CompletedAt: record.CompletedAt, IsPublished: record.IsPublished, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}
