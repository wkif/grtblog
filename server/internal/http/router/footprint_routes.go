package router

import (
	"github.com/gofiber/fiber/v2"
	appfootprint "github.com/grtsinry43/grtblog-v2/server/internal/app/footprint"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerFootprintRoutes(v2 fiber.Router, deps Dependencies) {
	h := handler.NewFootprintHandler(
		appfootprint.NewService(persistence.NewFootprintRepository(deps.DB), deps.EventBus),
		deps.SysConfig,
	)
	v2.Get("/footprints", h.ListPublic)

	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	requireAdmin := middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo)
	adminOnly := middleware.RequireAdmin(identityRepo)

	auth := v2.Group("/footprints", requireAdmin, adminOnly)
	auth.Post("/", h.Create)
	auth.Put("/:id", h.Update)
	auth.Delete("/:id", h.Delete)

	admin := v2.Group("/admin", requireAdmin, adminOnly)
	admin.Get("/footprints", h.ListAdmin)
	admin.Get("/footprints/:id", h.GetAdmin)
	admin.Get("/footprint-places", h.ListPlaces)
}
