package footprint

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/footprint"
)

type Service struct {
	repo   domain.Repository
	events appEvent.Bus
}

func NewService(repo domain.Repository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, events: events}
}

func (s *Service) Create(ctx context.Context, cmd CreateCmd) (*domain.Journey, error) {
	journey, place, albumIDs, err := normalizeCommand(cmd)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, journey, place, albumIDs); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, JourneyCreated{ID: journey.ID, At: time.Now()})
	return s.repo.GetByID(ctx, journey.ID, false)
}

func (s *Service) Update(ctx context.Context, cmd UpdateCmd) (*domain.Journey, error) {
	journey, place, albumIDs, err := normalizeCommand(cmd.CreateCmd)
	if err != nil {
		return nil, err
	}
	journey.ID = cmd.ID
	if err := s.repo.Update(ctx, journey, place, albumIDs); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, JourneyUpdated{ID: journey.ID, At: time.Now()})
	return s.repo.GetByID(ctx, journey.ID, false)
}

func (s *Service) Get(ctx context.Context, id int64) (*domain.Journey, error) {
	return s.repo.GetByID(ctx, id, false)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, JourneyDeleted{ID: id, At: time.Now()})
	return nil
}

func (s *Service) List(ctx context.Context, opts domain.ListOptions) ([]*domain.Journey, int64, error) {
	return s.repo.List(ctx, opts)
}

func (s *Service) ListPlaces(ctx context.Context) ([]domain.Place, error) {
	return s.repo.ListPlaces(ctx)
}

func (s *Service) PublicOverview(ctx context.Context) (*domain.Overview, error) {
	published := true
	journeys, _, err := s.repo.List(ctx, domain.ListOptions{
		Page:                1,
		PageSize:            1000,
		Published:           &published,
		PublishedAlbumsOnly: true,
	})
	if err != nil {
		return nil, err
	}

	overview := &domain.Overview{JourneyCount: len(journeys), Places: make([]domain.PlaceGroup, 0)}
	placeIndex := make(map[int64]int)
	for _, journey := range journeys {
		index, ok := placeIndex[journey.PlaceID]
		if !ok {
			index = len(overview.Places)
			placeIndex[journey.PlaceID] = index
			overview.Places = append(overview.Places, domain.PlaceGroup{Place: journey.Place})
		}
		group := &overview.Places[index]
		group.Journeys = append(group.Journeys, journey)
		group.Stats.JourneyCount++
		if journey.DistanceMeters != nil {
			group.Stats.TotalDistanceMeters += *journey.DistanceMeters
			overview.TotalDistanceMeters += *journey.DistanceMeters
		}
		if journey.DurationSeconds != nil {
			group.Stats.TotalDurationSeconds += *journey.DurationSeconds
			overview.TotalDurationSeconds += *journey.DurationSeconds
		}
	}
	overview.CityCount = len(overview.Places)
	sort.SliceStable(overview.Places, func(i, j int) bool {
		return overview.Places[i].Journeys[0].JourneyDate.After(overview.Places[j].Journeys[0].JourneyDate)
	})
	return overview, nil
}

func normalizeCommand(cmd CreateCmd) (*domain.Journey, domain.Place, []int64, error) {
	place := domain.Place{
		Slug:        strings.TrimSpace(cmd.Place.Slug),
		CityName:    strings.TrimSpace(cmd.Place.CityName),
		RegionName:  cleanString(cmd.Place.RegionName),
		CountryName: strings.TrimSpace(cmd.Place.CountryName),
		CountryCode: cleanUpperString(cmd.Place.CountryCode),
		Latitude:    cmd.Place.Latitude,
		Longitude:   cmd.Place.Longitude,
	}
	if place.Slug == "" || strings.ContainsAny(place.Slug, "/?#") || place.CityName == "" || place.CountryName == "" || cmd.JourneyDate.IsZero() {
		return nil, domain.Place{}, nil, domain.ErrInvalidInput
	}
	if place.Latitude < -90 || place.Latitude > 90 || place.Longitude < -180 || place.Longitude > 180 {
		return nil, domain.Place{}, nil, domain.ErrInvalidInput
	}
	if cmd.EndedAt != nil && cmd.EndedAt.Before(cmd.JourneyDate) {
		return nil, domain.Place{}, nil, domain.ErrInvalidInput
	}
	if (cmd.DistanceMeters != nil && *cmd.DistanceMeters < 0) || (cmd.DurationSeconds != nil && *cmd.DurationSeconds < 0) {
		return nil, domain.Place{}, nil, domain.ErrInvalidInput
	}
	trackURL := cleanString(cmd.TrackURL)
	if trackURL != nil {
		parsed, err := url.ParseRequestURI(*trackURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, domain.Place{}, nil, domain.ErrInvalidTrackURL
		}
	}
	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return nil, domain.Place{}, nil, domain.ErrInvalidInput
	}
	journey := &domain.Journey{
		Title:           title,
		JourneyDate:     dateOnly(cmd.JourneyDate),
		EndedAt:         dateOnlyPtr(cmd.EndedAt),
		Summary:         cleanString(cmd.Summary),
		Cover:           cleanString(cmd.Cover),
		DistanceMeters:  cmd.DistanceMeters,
		DurationSeconds: cmd.DurationSeconds,
		TrackURL:        trackURL,
		IsPublished:     cmd.IsPublished,
		SortOrder:       cmd.SortOrder,
	}
	return journey, place, uniquePositiveIDs(cmd.AlbumIDs), nil
}

func cleanString(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func cleanUpperString(value *string) *string {
	cleaned := cleanString(value)
	if cleaned == nil {
		return nil
	}
	upper := strings.ToUpper(*cleaned)
	return &upper
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func dateOnlyPtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	date := dateOnly(*value)
	return &date
}
