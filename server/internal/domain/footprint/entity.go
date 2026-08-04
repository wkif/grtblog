package footprint

import "time"

type Place struct {
	ID          int64
	Slug        string
	CityName    string
	RegionName  *string
	CountryName string
	CountryCode *string
	Latitude    float64
	Longitude   float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AlbumLink struct {
	ID          int64
	Title       string
	ShortURL    string
	Cover       *string
	PhotoCount  int64
	IsPublished bool
}

type Journey struct {
	ID              int64
	PlaceID         int64
	Place           Place
	Title           string
	JourneyDate     time.Time
	EndedAt         *time.Time
	Summary         *string
	Cover           *string
	DistanceMeters  *int64
	DurationSeconds *int64
	TrackURL        *string
	IsPublished     bool
	SortOrder       int
	Albums          []AlbumLink
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PlaceStats struct {
	JourneyCount         int
	TotalDistanceMeters  int64
	TotalDurationSeconds int64
}

type PlaceGroup struct {
	Place    Place
	Stats    PlaceStats
	Journeys []*Journey
}

type Overview struct {
	CityCount            int
	JourneyCount         int
	TotalDistanceMeters  int64
	TotalDurationSeconds int64
	Places               []PlaceGroup
}
