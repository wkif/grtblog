package contract

import "time"

type CreateMediaRecordReq struct {
	Title          string     `json:"title" validate:"required,max=255"`
	OriginalTitle  *string    `json:"originalTitle,omitempty"`
	MediaType      string     `json:"mediaType" validate:"required,oneof=movie tv"`
	Provider       string     `json:"provider"`
	ProviderID     *string    `json:"providerId,omitempty"`
	Poster         *string    `json:"poster,omitempty"`
	Backdrop       *string    `json:"backdrop,omitempty"`
	Overview       *string    `json:"overview,omitempty"`
	ReleaseDate    *time.Time `json:"releaseDate,omitempty"`
	RuntimeMinutes *int       `json:"runtimeMinutes,omitempty"`
	TotalEpisodes  *int       `json:"totalEpisodes,omitempty"`
	Status         string     `json:"status" validate:"required,oneof=planned watching completed dropped"`
	Progress       int        `json:"progress"`
	ProgressTotal  *int       `json:"progressTotal,omitempty"`
	Rating         *float64   `json:"rating,omitempty"`
	Note           *string    `json:"note,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	IsPublished    bool       `json:"isPublished"`
}

type UpdateMediaRecordReq = CreateMediaRecordReq
