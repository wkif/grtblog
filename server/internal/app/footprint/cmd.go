package footprint

import "time"

type PlaceInput struct {
	Slug        string
	CityName    string
	RegionName  *string
	CountryName string
	CountryCode *string
	Latitude    float64
	Longitude   float64
}

type CreateCmd struct {
	Place           PlaceInput
	Title           string
	JourneyDate     time.Time
	EndedAt         *time.Time
	Summary         *string
	Cover           *string
	DistanceMeters  *int64
	DurationSeconds *int64
	TrackURL        *string
	AlbumIDs        []int64
	IsPublished     bool
	SortOrder       int
}

type UpdateCmd struct {
	ID int64
	CreateCmd
}
