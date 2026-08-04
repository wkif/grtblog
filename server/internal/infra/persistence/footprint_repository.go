package persistence

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/footprint"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type FootprintRepository struct{ db *gorm.DB }

func NewFootprintRepository(db *gorm.DB) *FootprintRepository {
	return &FootprintRepository{db: db}
}

func (r *FootprintRepository) Create(ctx context.Context, journey *domain.Journey, place domain.Place, albumIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		placeModel, err := upsertFootprintPlace(tx, place)
		if err != nil {
			return err
		}
		journey.PlaceID = placeModel.ID
		journey.Place = toFootprintPlaceDomain(placeModel)
		modelJourney := toFootprintJourneyModel(journey)
		if err := tx.Create(&modelJourney).Error; err != nil {
			return err
		}
		journey.ID = modelJourney.ID
		journey.CreatedAt = modelJourney.CreatedAt
		journey.UpdatedAt = modelJourney.UpdatedAt
		if err := replaceFootprintAlbums(tx, journey.ID, albumIDs); err != nil {
			return err
		}
		return nil
	})
}

func (r *FootprintRepository) GetByID(ctx context.Context, id int64, publishedAlbumsOnly bool) (*domain.Journey, error) {
	var row model.FootprintJourney
	if err := r.db.WithContext(ctx).Preload("Place").First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrJourneyNotFound
		}
		return nil, err
	}
	journey := toFootprintJourneyDomain(&row)
	links, err := loadFootprintAlbumLinks(r.db.WithContext(ctx), []int64{id}, publishedAlbumsOnly)
	if err != nil {
		return nil, err
	}
	journey.Albums = links[id]
	return journey, nil
}

func (r *FootprintRepository) Update(ctx context.Context, journey *domain.Journey, place domain.Place, albumIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.FootprintJourney
		if err := tx.First(&existing, journey.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrJourneyNotFound
			}
			return err
		}
		placeModel, err := upsertFootprintPlace(tx, place)
		if err != nil {
			return err
		}
		updates := map[string]any{
			"place_id": placeModel.ID, "title": journey.Title, "journey_date": journey.JourneyDate,
			"ended_at": journey.EndedAt, "summary": journey.Summary, "cover": journey.Cover,
			"distance_meters": journey.DistanceMeters, "duration_seconds": journey.DurationSeconds,
			"track_url": journey.TrackURL, "is_published": journey.IsPublished, "sort_order": journey.SortOrder,
		}
		if err := tx.Model(&model.FootprintJourney{}).Where("id = ?", journey.ID).Updates(updates).Error; err != nil {
			return err
		}
		return replaceFootprintAlbums(tx, journey.ID, albumIDs)
	})
}

func (r *FootprintRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&model.FootprintJourney{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrJourneyNotFound
		}
		return tx.Where("journey_id = ?", id).Delete(&model.FootprintJourneyAlbum{}).Error
	})
}

func (r *FootprintRepository) List(ctx context.Context, opts domain.ListOptions) ([]*domain.Journey, int64, error) {
	page, size := opts.Page, opts.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 1000 {
		size = 1000
	}
	query := r.db.WithContext(ctx).Model(&model.FootprintJourney{}).
		Joins("JOIN footprint_place fp ON fp.id = footprint_journey.place_id AND fp.deleted_at IS NULL")
	if opts.Published != nil {
		query = query.Where("footprint_journey.is_published = ?", *opts.Published)
	}
	if search := strings.TrimSpace(opts.Search); search != "" {
		term := "%" + search + "%"
		query = query.Where("footprint_journey.title ILIKE ? OR fp.city_name ILIKE ? OR fp.region_name ILIKE ?", term, term, term)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.FootprintJourney
	if err := query.Preload("Place").Order("footprint_journey.journey_date DESC, footprint_journey.sort_order DESC, footprint_journey.id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	journeys := make([]*domain.Journey, len(rows))
	ids := make([]int64, len(rows))
	for i := range rows {
		journeys[i] = toFootprintJourneyDomain(&rows[i])
		ids[i] = rows[i].ID
	}
	links, err := loadFootprintAlbumLinks(r.db.WithContext(ctx), ids, opts.PublishedAlbumsOnly)
	if err != nil {
		return nil, 0, err
	}
	for _, journey := range journeys {
		journey.Albums = links[journey.ID]
	}
	return journeys, total, nil
}

func (r *FootprintRepository) ListPlaces(ctx context.Context) ([]domain.Place, error) {
	var rows []model.FootprintPlace
	if err := r.db.WithContext(ctx).Order("country_name, region_name, city_name").Find(&rows).Error; err != nil {
		return nil, err
	}
	places := make([]domain.Place, len(rows))
	for i := range rows {
		places[i] = toFootprintPlaceDomain(&rows[i])
	}
	return places, nil
}

func upsertFootprintPlace(tx *gorm.DB, place domain.Place) (*model.FootprintPlace, error) {
	var row model.FootprintPlace
	err := tx.Where("slug = ?", place.Slug).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = model.FootprintPlace{
			Slug: place.Slug, CityName: place.CityName, RegionName: place.RegionName,
			CountryName: place.CountryName, CountryCode: place.CountryCode,
			Latitude: place.Latitude, Longitude: place.Longitude,
		}
		if err := tx.Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Model(&row).Updates(map[string]any{
		"city_name": place.CityName, "region_name": place.RegionName, "country_name": place.CountryName,
		"country_code": place.CountryCode, "latitude": place.Latitude, "longitude": place.Longitude,
	}).Error; err != nil {
		return nil, err
	}
	row.CityName, row.RegionName, row.CountryName = place.CityName, place.RegionName, place.CountryName
	row.CountryCode, row.Latitude, row.Longitude = place.CountryCode, place.Latitude, place.Longitude
	return &row, nil
}

func replaceFootprintAlbums(tx *gorm.DB, journeyID int64, albumIDs []int64) error {
	if err := tx.Where("journey_id = ?", journeyID).Delete(&model.FootprintJourneyAlbum{}).Error; err != nil {
		return err
	}
	if len(albumIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&model.Album{}).Where("id IN ?", albumIDs).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(albumIDs)) {
		return domain.ErrAlbumNotFound
	}
	links := make([]model.FootprintJourneyAlbum, len(albumIDs))
	for i, albumID := range albumIDs {
		links[i] = model.FootprintJourneyAlbum{JourneyID: journeyID, AlbumID: albumID, SortOrder: i}
	}
	return tx.Create(&links).Error
}

type footprintAlbumRow struct {
	JourneyID   int64
	ID          int64
	Title       string
	ShortURL    string
	Cover       *string
	PhotoCount  int64
	IsPublished bool
	SortOrder   int
}

func loadFootprintAlbumLinks(db *gorm.DB, journeyIDs []int64, publishedOnly bool) (map[int64][]domain.AlbumLink, error) {
	result := make(map[int64][]domain.AlbumLink, len(journeyIDs))
	if len(journeyIDs) == 0 {
		return result, nil
	}
	query := db.Table("footprint_journey_album fja").
		Select("fja.journey_id, fja.sort_order, a.id, a.title, a.short_url, a.cover, a.is_published, COUNT(DISTINCT p.id) AS photo_count").
		Joins("JOIN album a ON a.id = fja.album_id AND a.deleted_at IS NULL").
		Joins("LEFT JOIN photo p ON p.album_id = a.id AND p.deleted_at IS NULL").
		Where("fja.journey_id IN ?", journeyIDs)
	if publishedOnly {
		query = query.Where("a.is_published = ?", true)
	}
	var rows []footprintAlbumRow
	if err := query.Group("fja.journey_id, fja.sort_order, a.id, a.title, a.short_url, a.cover, a.is_published").
		Order("fja.journey_id, fja.sort_order, a.id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.JourneyID] = append(result[row.JourneyID], domain.AlbumLink{
			ID: row.ID, Title: row.Title, ShortURL: row.ShortURL, Cover: row.Cover,
			PhotoCount: row.PhotoCount, IsPublished: row.IsPublished,
		})
	}
	return result, nil
}

func toFootprintJourneyModel(journey *domain.Journey) model.FootprintJourney {
	return model.FootprintJourney{
		ID: journey.ID, PlaceID: journey.PlaceID, Title: journey.Title, JourneyDate: journey.JourneyDate,
		EndedAt: journey.EndedAt, Summary: journey.Summary, Cover: journey.Cover,
		DistanceMeters: journey.DistanceMeters, DurationSeconds: journey.DurationSeconds,
		TrackURL: journey.TrackURL, IsPublished: journey.IsPublished, SortOrder: journey.SortOrder,
		CreatedAt: journey.CreatedAt, UpdatedAt: journey.UpdatedAt,
	}
}

func toFootprintJourneyDomain(row *model.FootprintJourney) *domain.Journey {
	return &domain.Journey{
		ID: row.ID, PlaceID: row.PlaceID, Place: toFootprintPlaceDomain(&row.Place), Title: row.Title,
		JourneyDate: row.JourneyDate, EndedAt: row.EndedAt, Summary: row.Summary, Cover: row.Cover,
		DistanceMeters: row.DistanceMeters, DurationSeconds: row.DurationSeconds, TrackURL: row.TrackURL,
		IsPublished: row.IsPublished, SortOrder: row.SortOrder, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func toFootprintPlaceDomain(row *model.FootprintPlace) domain.Place {
	return domain.Place{
		ID: row.ID, Slug: row.Slug, CityName: row.CityName, RegionName: row.RegionName,
		CountryName: row.CountryName, CountryCode: row.CountryCode,
		Latitude: row.Latitude, Longitude: row.Longitude, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
