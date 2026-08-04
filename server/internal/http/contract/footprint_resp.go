package contract

import (
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/footprint"
)

type FootprintPlaceResp struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	CityName    string     `json:"cityName"`
	RegionName  *string    `json:"regionName,omitempty"`
	CountryName string     `json:"countryName"`
	CountryCode *string    `json:"countryCode,omitempty"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type FootprintAlbumResp struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	ShortURL    string  `json:"shortUrl"`
	Cover       *string `json:"cover,omitempty"`
	PhotoCount  int64   `json:"photoCount"`
	IsPublished bool    `json:"isPublished"`
}

type FootprintJourneyResp struct {
	ID              int64                `json:"id"`
	PlaceID         int64                `json:"placeId"`
	Place           FootprintPlaceResp   `json:"place"`
	Title           string               `json:"title"`
	JourneyDate     time.Time            `json:"journeyDate"`
	EndedAt         *time.Time           `json:"endedAt,omitempty"`
	Summary         *string              `json:"summary,omitempty"`
	Cover           *string              `json:"cover,omitempty"`
	DistanceMeters  *int64               `json:"distanceMeters,omitempty"`
	DurationSeconds *int64               `json:"durationSeconds,omitempty"`
	TrackURL        *string              `json:"trackUrl,omitempty"`
	Albums          []FootprintAlbumResp `json:"albums"`
	IsPublished     bool                 `json:"isPublished"`
	SortOrder       int                  `json:"sortOrder"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

type FootprintJourneyListResp struct {
	Items []FootprintJourneyResp `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}

type FootprintStatsResp struct {
	JourneyCount         int   `json:"journeyCount"`
	TotalDistanceMeters  int64 `json:"totalDistanceMeters"`
	TotalDurationSeconds int64 `json:"totalDurationSeconds"`
}

type FootprintPlaceGroupResp struct {
	FootprintPlaceResp
	Stats    FootprintStatsResp     `json:"stats"`
	Journeys []FootprintJourneyResp `json:"journeys"`
}

type FootprintOverviewResp struct {
	Summary struct {
		CityCount            int   `json:"cityCount"`
		JourneyCount         int   `json:"journeyCount"`
		TotalDistanceMeters  int64 `json:"totalDistanceMeters"`
		TotalDurationSeconds int64 `json:"totalDurationSeconds"`
	} `json:"summary"`
	Places []FootprintPlaceGroupResp `json:"places"`
	Map    FootprintMapSettingsResp  `json:"map"`
}

type FootprintMapSettingsResp struct {
	Provider      string `json:"provider"`
	TiandituKey   string `json:"tiandituKey"`
	TiandituLayer string `json:"tiandituLayer"`
}

func FootprintPlaceResponse(place domain.Place, includeTimestamps bool) FootprintPlaceResp {
	response := FootprintPlaceResp{
		ID: place.ID, Slug: place.Slug, CityName: place.CityName, RegionName: place.RegionName,
		CountryName: place.CountryName, CountryCode: place.CountryCode,
		Latitude: place.Latitude, Longitude: place.Longitude,
	}
	if includeTimestamps {
		response.CreatedAt = &place.CreatedAt
		response.UpdatedAt = &place.UpdatedAt
	}
	return response
}

func FootprintJourneyResponse(journey *domain.Journey) FootprintJourneyResp {
	albums := make([]FootprintAlbumResp, len(journey.Albums))
	for i, album := range journey.Albums {
		albums[i] = FootprintAlbumResp{
			ID: album.ID, Title: album.Title, ShortURL: album.ShortURL, Cover: album.Cover,
			PhotoCount: album.PhotoCount, IsPublished: album.IsPublished,
		}
	}
	return FootprintJourneyResp{
		ID: journey.ID, PlaceID: journey.PlaceID, Place: FootprintPlaceResponse(journey.Place, false),
		Title: journey.Title, JourneyDate: journey.JourneyDate, EndedAt: journey.EndedAt,
		Summary: journey.Summary, Cover: journey.Cover, DistanceMeters: journey.DistanceMeters,
		DurationSeconds: journey.DurationSeconds, TrackURL: journey.TrackURL, Albums: albums,
		IsPublished: journey.IsPublished, SortOrder: journey.SortOrder,
		CreatedAt: journey.CreatedAt, UpdatedAt: journey.UpdatedAt,
	}
}

func FootprintOverviewResponse(overview *domain.Overview, mapSettings sysconfig.MapSettings) FootprintOverviewResp {
	response := FootprintOverviewResp{
		Places: make([]FootprintPlaceGroupResp, len(overview.Places)),
		Map: FootprintMapSettingsResp{
			Provider:      mapSettings.Provider,
			TiandituKey:   mapSettings.TiandituKey,
			TiandituLayer: mapSettings.TiandituLayer,
		},
	}
	response.Summary.CityCount = overview.CityCount
	response.Summary.JourneyCount = overview.JourneyCount
	response.Summary.TotalDistanceMeters = overview.TotalDistanceMeters
	response.Summary.TotalDurationSeconds = overview.TotalDurationSeconds
	for i, group := range overview.Places {
		journeys := make([]FootprintJourneyResp, len(group.Journeys))
		for j, journey := range group.Journeys {
			journeys[j] = FootprintJourneyResponse(journey)
		}
		response.Places[i] = FootprintPlaceGroupResp{
			FootprintPlaceResp: FootprintPlaceResponse(group.Place, false),
			Stats: FootprintStatsResp{
				JourneyCount: group.Stats.JourneyCount, TotalDistanceMeters: group.Stats.TotalDistanceMeters,
				TotalDurationSeconds: group.Stats.TotalDurationSeconds,
			},
			Journeys: journeys,
		}
	}
	return response
}
