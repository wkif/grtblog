package mediarecord

import (
	"context"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/mediarecord"
)

type SearchProvider interface {
	Search(ctx context.Context, query string, mediaType string) ([]domain.SearchResult, error)
}

type DetailProvider interface {
	Details(ctx context.Context, providerID string, mediaType string) (domain.SearchResult, error)
}

type Service struct {
	repo     domain.Repository
	provider SearchProvider
	events   appEvent.Bus
}

func NewService(repo domain.Repository, provider SearchProvider) *Service {
	return NewServiceWithEvents(repo, provider, nil)
}

func NewServiceWithEvents(repo domain.Repository, provider SearchProvider, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, provider: provider, events: events}
}

func (s *Service) Create(ctx context.Context, cmd CreateCmd) (*domain.Record, error) {
	if err := validate(cmd.MediaType, cmd.Status); err != nil {
		return nil, err
	}
	record := &domain.Record{
		Title: cmd.Title, OriginalTitle: cmd.OriginalTitle, MediaType: cmd.MediaType,
		Provider: cmd.Provider, ProviderID: cmd.ProviderID, Poster: cmd.Poster, Backdrop: cmd.Backdrop,
		Overview: cmd.Overview, ReleaseDate: cmd.ReleaseDate, RuntimeMinutes: cmd.RuntimeMinutes,
		TotalEpisodes: cmd.TotalEpisodes, Status: cmd.Status, Progress: cmd.Progress,
		ProgressTotal: cmd.ProgressTotal, Rating: cmd.Rating, Note: cmd.Note, StartedAt: cmd.StartedAt,
		CompletedAt: cmd.CompletedAt, IsPublished: cmd.IsPublished,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, MediaRecordCreated{ID: record.ID, Published: record.IsPublished, At: time.Now()})
	return record, nil
}

func (s *Service) Update(ctx context.Context, cmd UpdateCmd) (*domain.Record, error) {
	if err := validate(cmd.MediaType, cmd.Status); err != nil {
		return nil, err
	}
	record, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	record.Title, record.OriginalTitle, record.MediaType = cmd.Title, cmd.OriginalTitle, cmd.MediaType
	record.Provider, record.ProviderID, record.Poster, record.Backdrop = cmd.Provider, cmd.ProviderID, cmd.Poster, cmd.Backdrop
	record.Overview, record.ReleaseDate, record.RuntimeMinutes = cmd.Overview, cmd.ReleaseDate, cmd.RuntimeMinutes
	record.TotalEpisodes, record.Status, record.Progress = cmd.TotalEpisodes, cmd.Status, cmd.Progress
	record.ProgressTotal, record.Rating, record.Note = cmd.ProgressTotal, cmd.Rating, cmd.Note
	record.StartedAt, record.CompletedAt, record.IsPublished = cmd.StartedAt, cmd.CompletedAt, cmd.IsPublished
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, MediaRecordUpdated{ID: record.ID, Published: record.IsPublished, At: time.Now()})
	return record, nil
}

func (s *Service) Get(ctx context.Context, id int64) (*domain.Record, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, MediaRecordDeleted{ID: id, At: time.Now()})
	return nil
}

func (s *Service) List(ctx context.Context, opts domain.ListOptions) ([]*domain.Record, int64, error) {
	return s.repo.List(ctx, opts)
}

func (s *Service) Search(ctx context.Context, query, mediaType string) ([]domain.SearchResult, error) {
	if s.provider == nil || strings.TrimSpace(query) == "" {
		return []domain.SearchResult{}, nil
	}
	return s.provider.Search(ctx, strings.TrimSpace(query), mediaType)
}

func (s *Service) Details(ctx context.Context, providerID, mediaType string) (domain.SearchResult, error) {
	provider, ok := s.provider.(DetailProvider)
	if !ok || strings.TrimSpace(providerID) == "" {
		return domain.SearchResult{}, domain.ErrNotFound
	}
	if mediaType != domain.TypeMovie && mediaType != domain.TypeTV {
		return domain.SearchResult{}, domain.ErrInvalidType
	}
	return provider.Details(ctx, strings.TrimSpace(providerID), mediaType)
}

func validate(mediaType, status string) error {
	if mediaType != domain.TypeMovie && mediaType != domain.TypeTV {
		return domain.ErrInvalidType
	}
	switch status {
	case domain.StatusPlanned, domain.StatusWatching, domain.StatusCompleted, domain.StatusDropped:
		return nil
	default:
		return domain.ErrInvalidStatus
	}
}
