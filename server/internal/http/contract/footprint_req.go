package contract

import "time"

type FootprintPlaceReq struct {
	Slug        string  `json:"slug"`
	CityName    string  `json:"cityName"`
	RegionName  *string `json:"regionName,omitempty"`
	CountryName string  `json:"countryName"`
	CountryCode *string `json:"countryCode,omitempty"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type CreateFootprintJourneyReq struct {
	Place           FootprintPlaceReq `json:"place"`
	Title           string            `json:"title"`
	JourneyDate     time.Time         `json:"journeyDate"`
	EndedAt         *time.Time        `json:"endedAt,omitempty"`
	Summary         *string           `json:"summary,omitempty"`
	Cover           *string           `json:"cover,omitempty"`
	DistanceMeters  *int64            `json:"distanceMeters,omitempty"`
	DurationSeconds *int64            `json:"durationSeconds,omitempty"`
	TrackURL        *string           `json:"trackUrl,omitempty"`
	AlbumIDs        []int64           `json:"albumIds"`
	IsPublished     bool              `json:"isPublished"`
	SortOrder       int               `json:"sortOrder"`
}

type UpdateFootprintJourneyReq = CreateFootprintJourneyReq
