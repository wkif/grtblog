package router

import (
	"context"

	"github.com/gofiber/fiber/v2"
	appmedia "github.com/grtsinry43/grtblog-v2/server/internal/app/mediarecord"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/mediarecord"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerMediaRecordRoutes(v2 fiber.Router, deps Dependencies) {
	h := newMediaRecordHandler(deps)
	public := v2.Group("/media-records")
	public.Get("/", h.ListPublic)
	public.Get("/:id", h.GetPublic)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	auth := v2.Group("/media-records", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	auth.Post("/", h.Create)
	auth.Put("/:id", h.Update)
	auth.Delete("/:id", h.Delete)
	admin := v2.Group("/admin", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	admin.Get("/media-records/search", h.Search)
	admin.Get("/media-records/details/:mediaType/:providerID", h.Details)
	admin.Get("/media-records", h.ListAdmin)
	admin.Get("/media-records/:id", h.GetAdmin)
}

func newMediaRecordHandler(deps Dependencies) *handler.MediaRecordHandler {
	repo := persistence.NewMediaRecordRepository(deps.DB)
	provider := mediarecord.NewDynamicTMDBClient(func(ctx context.Context) mediarecord.TMDBSettings {
		settings := deps.SysConfig.MediaSettings(ctx)
		return mediarecord.TMDBSettings{
			APIKey:       settings.TMDBAPIKey,
			Language:     settings.TMDBLanguage,
			BaseURL:      settings.TMDBBaseURL,
			ImageBaseURL: settings.TMDBImageBaseURL,
			Timeout:      settings.TMDBTimeout,
		}
	})
	return handler.NewMediaRecordHandler(appmedia.NewServiceWithEvents(repo, provider, deps.EventBus))
}
