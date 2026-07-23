package model

import (
	"time"

	"gorm.io/gorm"
)

type MediaRecord struct {
	ID             int64          `gorm:"column:id;primaryKey"`
	Title          string         `gorm:"column:title;size:255;not null"`
	OriginalTitle  *string        `gorm:"column:original_title;size:255"`
	MediaType      string         `gorm:"column:media_type;size:16;not null"`
	Provider       string         `gorm:"column:provider;size:32;not null"`
	ProviderID     *string        `gorm:"column:provider_id;size:64"`
	Poster         *string        `gorm:"column:poster;size:1024"`
	Backdrop       *string        `gorm:"column:backdrop;size:1024"`
	Overview       *string        `gorm:"column:overview;type:text"`
	ReleaseDate    *time.Time     `gorm:"column:release_date;type:date"`
	RuntimeMinutes *int           `gorm:"column:runtime_minutes"`
	TotalEpisodes  *int           `gorm:"column:total_episodes"`
	Status         string         `gorm:"column:status;size:16;not null"`
	Progress       int            `gorm:"column:progress;not null"`
	ProgressTotal  *int           `gorm:"column:progress_total"`
	Rating         *float64       `gorm:"column:rating;type:numeric(3,1)"`
	Note           *string        `gorm:"column:note;type:text"`
	StartedAt      *time.Time     `gorm:"column:started_at"`
	CompletedAt    *time.Time     `gorm:"column:completed_at"`
	IsPublished    bool           `gorm:"column:is_published;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (MediaRecord) TableName() string { return "media_record" }
