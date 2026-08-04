package footprint

import (
	"errors"
	"testing"
	"time"

	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/footprint"
)

func TestNormalizeCommand(t *testing.T) {
	date := time.Date(2026, time.July, 18, 14, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	endedAt := date.Add(24 * time.Hour)
	region := " 内蒙古自治区 "
	countryCode := " cn "
	trackURL := " https://example.com/tracks/tengger "
	distance := int64(27800)
	duration := int64(8*60*60 + 25*60)

	journey, place, albumIDs, err := normalizeCommand(CreateCmd{
		Place: PlaceInput{
			Slug:        " alxa-left-banner ",
			CityName:    " 阿拉善左旗 ",
			RegionName:  &region,
			CountryName: " 中国 ",
			CountryCode: &countryCode,
			Latitude:    38.833,
			Longitude:   105.668,
		},
		Title:           " 腾格里沙漠徒步 ",
		JourneyDate:     date,
		EndedAt:         &endedAt,
		DistanceMeters:  &distance,
		DurationSeconds: &duration,
		TrackURL:        &trackURL,
		AlbumIDs:        []int64{3, 3, -1, 8},
		IsPublished:     true,
	})
	if err != nil {
		t.Fatalf("normalizeCommand returned error: %v", err)
	}
	if journey.Title != "腾格里沙漠徒步" || journey.TrackURL == nil || *journey.TrackURL != "https://example.com/tracks/tengger" {
		t.Fatalf("journey was not normalized: %#v", journey)
	}
	if !journey.JourneyDate.Equal(time.Date(2026, time.July, 18, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("journey date was not normalized to UTC date: %v", journey.JourneyDate)
	}
	if place.Slug != "alxa-left-banner" || place.CityName != "阿拉善左旗" || place.CountryCode == nil || *place.CountryCode != "CN" {
		t.Fatalf("place was not normalized: %#v", place)
	}
	if len(albumIDs) != 2 || albumIDs[0] != 3 || albumIDs[1] != 8 {
		t.Fatalf("album IDs were not deduplicated: %#v", albumIDs)
	}
}

func TestNormalizeCommandRejectsInvalidOptionalMetrics(t *testing.T) {
	date := time.Date(2026, time.July, 18, 0, 0, 0, 0, time.UTC)
	valid := CreateCmd{
		Place: PlaceInput{
			Slug:        "alxa-left-banner",
			CityName:    "阿拉善左旗",
			CountryName: "中国",
			Latitude:    38.833,
			Longitude:   105.668,
		},
		Title:       "腾格里沙漠徒步",
		JourneyDate: date,
	}

	tests := []struct {
		name string
		edit func(*CreateCmd)
		want error
	}{
		{
			name: "invalid track URL",
			edit: func(cmd *CreateCmd) {
				value := "ftp://example.com/track.gpx"
				cmd.TrackURL = &value
			},
			want: domain.ErrInvalidTrackURL,
		},
		{
			name: "negative distance",
			edit: func(cmd *CreateCmd) {
				value := int64(-1)
				cmd.DistanceMeters = &value
			},
			want: domain.ErrInvalidInput,
		},
		{
			name: "end before start",
			edit: func(cmd *CreateCmd) {
				value := date.Add(-24 * time.Hour)
				cmd.EndedAt = &value
			},
			want: domain.ErrInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := valid
			test.edit(&cmd)
			_, _, _, err := normalizeCommand(cmd)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}
