package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	appfootprint "github.com/grtsinry43/grtblog-v2/server/internal/app/footprint"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/footprint"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type FootprintHandler struct {
	svc       *appfootprint.Service
	sysConfig *sysconfig.Service
}

func NewFootprintHandler(svc *appfootprint.Service, sysConfig *sysconfig.Service) *FootprintHandler {
	return &FootprintHandler{svc: svc, sysConfig: sysConfig}
}

// ListPublic godoc
// @Summary 获取公开足迹总览
// @Tags Footprint
// @Produce json
// @Success 200 {object} contract.FootprintOverviewResp
// @Router /footprints [get]
func (h *FootprintHandler) ListPublic(c *fiber.Ctx) error {
	overview, err := h.svc.PublicOverview(c.Context())
	if err != nil {
		return err
	}
	mapSettings := sysconfig.MapSettings{Provider: "osm", TiandituLayer: "vector"}
	if h.sysConfig != nil {
		mapSettings, err = h.sysConfig.MapSettings(c.Context())
		if err != nil {
			return err
		}
	}
	return response.Success(c, contract.FootprintOverviewResponse(overview, mapSettings))
}

// ListAdmin godoc
// @Summary 获取足迹行程列表（管理端）
// @Tags Footprint
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param search query string false "行程或城市关键词"
// @Param published query bool false "发布状态"
// @Success 200 {object} contract.FootprintJourneyListResp
// @Security BearerAuth
// @Router /admin/footprints [get]
func (h *FootprintHandler) ListAdmin(c *fiber.Ctx) error {
	page, size := queryPage(c)
	var published *bool
	if raw := strings.TrimSpace(c.Query("published")); raw != "" {
		value := raw == "true"
		published = &value
	}
	items, total, err := h.svc.List(c.Context(), domain.ListOptions{
		Page: page, PageSize: size, Published: published, Search: c.Query("search"),
	})
	if err != nil {
		return err
	}
	output := make([]contract.FootprintJourneyResp, len(items))
	for i, item := range items {
		output[i] = contract.FootprintJourneyResponse(item)
	}
	return response.Success(c, contract.FootprintJourneyListResp{Items: output, Total: total, Page: page, Size: size})
}

// GetAdmin godoc
// @Summary 获取足迹行程详情（管理端）
// @Tags Footprint
// @Produce json
// @Param id path int true "足迹行程ID"
// @Success 200 {object} contract.FootprintJourneyResp
// @Security BearerAuth
// @Router /admin/footprints/{id} [get]
func (h *FootprintHandler) GetAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ErrorWithMsg[any](c, response.ParamsError, "无效的足迹记录ID")
	}
	journey, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return footprintError(c, err)
	}
	return response.Success(c, contract.FootprintJourneyResponse(journey))
}

// ListPlaces godoc
// @Summary 获取足迹城市列表（管理端）
// @Tags Footprint
// @Produce json
// @Success 200 {object} []contract.FootprintPlaceResp
// @Security BearerAuth
// @Router /admin/footprint-places [get]
func (h *FootprintHandler) ListPlaces(c *fiber.Ctx) error {
	places, err := h.svc.ListPlaces(c.Context())
	if err != nil {
		return err
	}
	output := make([]contract.FootprintPlaceResp, len(places))
	for i, place := range places {
		output[i] = contract.FootprintPlaceResponse(place, true)
	}
	return response.Success(c, output)
}

// Create godoc
// @Summary 创建足迹行程
// @Tags Footprint
// @Accept json
// @Produce json
// @Param request body contract.CreateFootprintJourneyReq true "足迹行程参数"
// @Success 200 {object} contract.FootprintJourneyResp
// @Security BearerAuth
// @Router /footprints [post]
func (h *FootprintHandler) Create(c *fiber.Ctx) error {
	var req contract.CreateFootprintJourneyReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	journey, err := h.svc.Create(c.Context(), footprintCreateCommand(req))
	if err != nil {
		return footprintError(c, err)
	}
	Audit(c, "footprint.create", map[string]any{"footprintId": journey.ID, "title": journey.Title})
	return response.SuccessWithMessage(c, contract.FootprintJourneyResponse(journey), "足迹记录创建成功")
}

// Update godoc
// @Summary 更新足迹行程
// @Tags Footprint
// @Accept json
// @Produce json
// @Param id path int true "足迹行程ID"
// @Param request body contract.UpdateFootprintJourneyReq true "足迹行程参数"
// @Success 200 {object} contract.FootprintJourneyResp
// @Security BearerAuth
// @Router /footprints/{id} [put]
func (h *FootprintHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ErrorWithMsg[any](c, response.ParamsError, "无效的足迹记录ID")
	}
	var req contract.UpdateFootprintJourneyReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	journey, err := h.svc.Update(c.Context(), appfootprint.UpdateCmd{ID: id, CreateCmd: footprintCreateCommand(req)})
	if err != nil {
		return footprintError(c, err)
	}
	return response.SuccessWithMessage(c, contract.FootprintJourneyResponse(journey), "足迹记录更新成功")
}

// Delete godoc
// @Summary 删除足迹行程
// @Tags Footprint
// @Produce json
// @Param id path int true "足迹行程ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /footprints/{id} [delete]
func (h *FootprintHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ErrorWithMsg[any](c, response.ParamsError, "无效的足迹记录ID")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return footprintError(c, err)
	}
	return response.SuccessWithMessage[any](c, nil, "足迹记录删除成功")
}

func footprintCreateCommand(req contract.CreateFootprintJourneyReq) appfootprint.CreateCmd {
	return appfootprint.CreateCmd{
		Place: appfootprint.PlaceInput{
			Slug: req.Place.Slug, CityName: req.Place.CityName, RegionName: req.Place.RegionName,
			CountryName: req.Place.CountryName, CountryCode: req.Place.CountryCode,
			Latitude: req.Place.Latitude, Longitude: req.Place.Longitude,
		},
		Title: req.Title, JourneyDate: req.JourneyDate, EndedAt: req.EndedAt, Summary: req.Summary,
		Cover: req.Cover, DistanceMeters: req.DistanceMeters, DurationSeconds: req.DurationSeconds,
		TrackURL: req.TrackURL, AlbumIDs: req.AlbumIDs, IsPublished: req.IsPublished, SortOrder: req.SortOrder,
	}
}

func footprintError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrJourneyNotFound):
		return response.ErrorWithMsg[any](c, response.NotFound, "足迹记录不存在")
	case errors.Is(err, domain.ErrAlbumNotFound):
		return response.ErrorWithMsg[any](c, response.ParamsError, "关联的相册不存在")
	case errors.Is(err, domain.ErrInvalidTrackURL):
		return response.ErrorWithMsg[any](c, response.ParamsError, "轨迹链接必须是有效的 HTTP(S) 地址")
	case errors.Is(err, domain.ErrInvalidInput):
		return response.ErrorWithMsg[any](c, response.ParamsError, "足迹记录参数无效")
	default:
		return err
	}
}
