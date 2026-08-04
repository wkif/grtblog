package model

import (
	"time"

	"gorm.io/gorm"
)

type FootprintPlace struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	Slug        string         `gorm:"column:slug;size:128;not null"`
	CityName    string         `gorm:"column:city_name;size:128;not null"`
	RegionName  *string        `gorm:"column:region_name;size:128"`
	CountryName string         `gorm:"column:country_name;size:128;not null"`
	CountryCode *string        `gorm:"column:country_code;size:8"`
	Latitude    float64        `gorm:"column:latitude;type:numeric(9,6);not null"`
	Longitude   float64        `gorm:"column:longitude;type:numeric(9,6);not null"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (FootprintPlace) TableName() string { return "footprint_place" }

type FootprintJourney struct {
	ID              int64          `gorm:"column:id;primaryKey"`
	PlaceID         int64          `gorm:"column:place_id;not null"`
	Place           FootprintPlace `gorm:"foreignKey:PlaceID"`
	Title           string         `gorm:"column:title;size:255;not null"`
	JourneyDate     time.Time      `gorm:"column:journey_date;type:date;not null"`
	EndedAt         *time.Time     `gorm:"column:ended_at;type:date"`
	Summary         *string        `gorm:"column:summary;type:text"`
	Cover           *string        `gorm:"column:cover;size:1024"`
	DistanceMeters  *int64         `gorm:"column:distance_meters"`
	DurationSeconds *int64         `gorm:"column:duration_seconds"`
	TrackURL        *string        `gorm:"column:track_url;size:2048"`
	IsPublished     bool           `gorm:"column:is_published;not null"`
	SortOrder       int            `gorm:"column:sort_order;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (FootprintJourney) TableName() string { return "footprint_journey" }

type FootprintJourneyAlbum struct {
	JourneyID int64 `gorm:"column:journey_id;primaryKey"`
	AlbumID   int64 `gorm:"column:album_id;primaryKey"`
	SortOrder int   `gorm:"column:sort_order;not null"`
}

func (FootprintJourneyAlbum) TableName() string { return "footprint_journey_album" }
