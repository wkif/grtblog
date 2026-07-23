package persistence

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/mediarecord"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type MediaRecordRepository struct{ db *gorm.DB }

func NewMediaRecordRepository(db *gorm.DB) *MediaRecordRepository {
	return &MediaRecordRepository{db: db}
}

func (r *MediaRecordRepository) Create(ctx context.Context, record *domain.Record) error {
	m := toMediaRecordModel(record)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	copyMediaRecord(record, &m)
	return nil
}

func (r *MediaRecordRepository) GetByID(ctx context.Context, id int64) (*domain.Record, error) {
	var m model.MediaRecord
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	d := toMediaRecordDomain(&m)
	return &d, nil
}

func (r *MediaRecordRepository) Update(ctx context.Context, record *domain.Record) error {
	m := toMediaRecordModel(record)
	if err := r.db.WithContext(ctx).Model(&model.MediaRecord{}).Where("id = ?", record.ID).Updates(map[string]any{
		"title": m.Title, "original_title": m.OriginalTitle, "media_type": m.MediaType, "provider": m.Provider,
		"provider_id": m.ProviderID, "poster": m.Poster, "backdrop": m.Backdrop, "overview": m.Overview,
		"release_date": m.ReleaseDate, "runtime_minutes": m.RuntimeMinutes, "total_episodes": m.TotalEpisodes,
		"status": m.Status, "progress": m.Progress, "progress_total": m.ProgressTotal, "rating": m.Rating,
		"note": m.Note, "started_at": m.StartedAt, "completed_at": m.CompletedAt, "is_published": m.IsPublished,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *MediaRecordRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.MediaRecord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MediaRecordRepository) List(ctx context.Context, opts domain.ListOptions) ([]*domain.Record, int64, error) {
	page, size := opts.Page, opts.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&model.MediaRecord{})
	if opts.Published != nil {
		q = q.Where("is_published = ?", *opts.Published)
	}
	if opts.Status != "" {
		q = q.Where("status = ?", opts.Status)
	}
	if opts.MediaType != "" {
		q = q.Where("media_type = ?", opts.MediaType)
	}
	if search := strings.TrimSpace(opts.SearchTerm); search != "" {
		q = q.Where("title ILIKE ? OR original_title ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.MediaRecord
	if err := q.Order("updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*domain.Record, len(rows))
	for i := range rows {
		d := toMediaRecordDomain(&rows[i])
		result[i] = &d
	}
	return result, total, nil
}

func toMediaRecordModel(d *domain.Record) model.MediaRecord {
	return model.MediaRecord{ID: d.ID, Title: d.Title, OriginalTitle: d.OriginalTitle, MediaType: d.MediaType, Provider: d.Provider, ProviderID: d.ProviderID, Poster: d.Poster, Backdrop: d.Backdrop, Overview: d.Overview, ReleaseDate: d.ReleaseDate, RuntimeMinutes: d.RuntimeMinutes, TotalEpisodes: d.TotalEpisodes, Status: d.Status, Progress: d.Progress, ProgressTotal: d.ProgressTotal, Rating: d.Rating, Note: d.Note, StartedAt: d.StartedAt, CompletedAt: d.CompletedAt, IsPublished: d.IsPublished, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
}

func toMediaRecordDomain(m *model.MediaRecord) domain.Record {
	return domain.Record{ID: m.ID, Title: m.Title, OriginalTitle: m.OriginalTitle, MediaType: m.MediaType, Provider: m.Provider, ProviderID: m.ProviderID, Poster: m.Poster, Backdrop: m.Backdrop, Overview: m.Overview, ReleaseDate: m.ReleaseDate, RuntimeMinutes: m.RuntimeMinutes, TotalEpisodes: m.TotalEpisodes, Status: m.Status, Progress: m.Progress, ProgressTotal: m.ProgressTotal, Rating: m.Rating, Note: m.Note, StartedAt: m.StartedAt, CompletedAt: m.CompletedAt, IsPublished: m.IsPublished, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func copyMediaRecord(d *domain.Record, m *model.MediaRecord) {
	d.ID, d.CreatedAt, d.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
}
