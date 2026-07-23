package handler

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	appmedia "github.com/grtsinry43/grtblog-v2/server/internal/app/mediarecord"
	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/mediarecord"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type MediaRecordHandler struct{ svc *appmedia.Service }

func NewMediaRecordHandler(svc *appmedia.Service) *MediaRecordHandler {
	return &MediaRecordHandler{svc: svc}
}

func (h *MediaRecordHandler) ListPublic(c *fiber.Ctx) error {
	p, size := queryPage(c)
	published := true
	items, total, err := h.svc.List(c.Context(), domain.ListOptions{Page: p, PageSize: size, Published: &published, Status: c.Query("status"), MediaType: c.Query("mediaType")})
	if err != nil {
		return err
	}
	return response.Success(c, buildMediaRecordList(items, total, p, size))
}

func (h *MediaRecordHandler) GetPublic(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ErrorWithMsg[any](c, response.NotFound, "影视记录不存在")
	}
	record, err := h.svc.Get(c.Context(), id)
	if errors.Is(err, domain.ErrNotFound) || (err == nil && !record.IsPublished) {
		return response.ErrorWithMsg[any](c, response.NotFound, "影视记录不存在")
	}
	if err != nil {
		return err
	}
	return response.Success(c, contract.MediaRecordResponse(record))
}

func (h *MediaRecordHandler) ListAdmin(c *fiber.Ctx) error {
	p, size := queryPage(c)
	var published *bool
	if value := strings.TrimSpace(c.Query("published")); value != "" {
		v := value == "true"
		published = &v
	}
	items, total, err := h.svc.List(c.Context(), domain.ListOptions{Page: p, PageSize: size, Published: published, Status: c.Query("status"), MediaType: c.Query("mediaType"), SearchTerm: c.Query("search")})
	if err != nil {
		return err
	}
	return response.Success(c, buildMediaRecordList(items, total, p, size))
}

func (h *MediaRecordHandler) GetAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ErrorWithMsg[any](c, response.ParamsError, "无效的影视记录ID")
	}
	record, err := h.svc.Get(c.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return response.ErrorWithMsg[any](c, response.NotFound, "影视记录不存在")
	}
	if err != nil {
		return err
	}
	return response.Success(c, contract.MediaRecordResponse(record))
}

func (h *MediaRecordHandler) Create(c *fiber.Ctx) error {
	var req contract.CreateMediaRecordReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	record, err := h.svc.Create(c.Context(), createCommand(req))
	if err != nil {
		return mediaRecordError(c, err)
	}
	Audit(c, "media_record.create", map[string]any{"mediaRecordId": record.ID, "title": record.Title})
	return response.SuccessWithMessage(c, contract.MediaRecordResponse(record), "影视记录创建成功")
}

func (h *MediaRecordHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ErrorWithMsg[any](c, response.ParamsError, "无效的影视记录ID")
	}
	var req contract.UpdateMediaRecordReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	cmd := createCommand(req)
	record, err := h.svc.Update(c.Context(), appmedia.UpdateCmd{ID: id, Title: cmd.Title, OriginalTitle: cmd.OriginalTitle, MediaType: cmd.MediaType, Provider: cmd.Provider, ProviderID: cmd.ProviderID, Poster: cmd.Poster, Backdrop: cmd.Backdrop, Overview: cmd.Overview, ReleaseDate: cmd.ReleaseDate, RuntimeMinutes: cmd.RuntimeMinutes, TotalEpisodes: cmd.TotalEpisodes, Status: cmd.Status, Progress: cmd.Progress, ProgressTotal: cmd.ProgressTotal, Rating: cmd.Rating, Note: cmd.Note, StartedAt: cmd.StartedAt, CompletedAt: cmd.CompletedAt, IsPublished: cmd.IsPublished})
	if err != nil {
		return mediaRecordError(c, err)
	}
	return response.SuccessWithMessage(c, contract.MediaRecordResponse(record), "影视记录更新成功")
}

func (h *MediaRecordHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ErrorWithMsg[any](c, response.ParamsError, "无效的影视记录ID")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return mediaRecordError(c, err)
	}
	return response.SuccessWithMessage[any](c, nil, "影视记录删除成功")
}

func (h *MediaRecordHandler) Search(c *fiber.Ctx) error {
	results, err := h.svc.Search(c.Context(), c.Query("q"), c.Query("mediaType"))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return response.NewBizErrorWithCause(response.ServerError, "TMDB 请求超时，请检查服务器网络或配置可访问的 API 地址", err)
		}
		return response.NewBizErrorWithCause(response.ServerError, "影视搜索失败，请检查 TMDB 配置和服务器网络", err)
	}
	output := make([]contract.MediaSearchResultResp, len(results))
	for i := range results {
		output[i] = contract.MediaSearchResultResp{ProviderID: results[i].ProviderID, Title: results[i].Title, OriginalTitle: results[i].OriginalTitle, MediaType: results[i].MediaType, Poster: results[i].Poster, Backdrop: results[i].Backdrop, Overview: results[i].Overview, ReleaseDate: results[i].ReleaseDate, RuntimeMinutes: results[i].RuntimeMinutes, TotalEpisodes: results[i].TotalEpisodes}
	}
	return response.Success(c, output)
}

func (h *MediaRecordHandler) Details(c *fiber.Ctx) error {
	result, err := h.svc.Details(c.Context(), c.Params("providerID"), c.Params("mediaType"))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return response.NewBizErrorWithCause(response.ServerError, "TMDB 详情请求超时", err)
		}
		return mediaRecordError(c, err)
	}
	return response.Success(c, mediaSearchResultResponse(result))
}

func mediaSearchResultResponse(result domain.SearchResult) contract.MediaSearchResultResp {
	return contract.MediaSearchResultResp{ProviderID: result.ProviderID, Title: result.Title, OriginalTitle: result.OriginalTitle, MediaType: result.MediaType, Poster: result.Poster, Backdrop: result.Backdrop, Overview: result.Overview, ReleaseDate: result.ReleaseDate, RuntimeMinutes: result.RuntimeMinutes, TotalEpisodes: result.TotalEpisodes}
}

func createCommand(req contract.CreateMediaRecordReq) appmedia.CreateCmd {
	return appmedia.CreateCmd{Title: req.Title, OriginalTitle: req.OriginalTitle, MediaType: req.MediaType, Provider: req.Provider, ProviderID: req.ProviderID, Poster: req.Poster, Backdrop: req.Backdrop, Overview: req.Overview, ReleaseDate: req.ReleaseDate, RuntimeMinutes: req.RuntimeMinutes, TotalEpisodes: req.TotalEpisodes, Status: req.Status, Progress: req.Progress, ProgressTotal: req.ProgressTotal, Rating: req.Rating, Note: req.Note, StartedAt: req.StartedAt, CompletedAt: req.CompletedAt, IsPublished: req.IsPublished}
}

func queryPage(c *fiber.Ctx) (int, int) {
	p, _ := strconv.Atoi(c.Query("page", "1"))
	size, _ := strconv.Atoi(c.Query("pageSize", "20"))
	if p < 1 {
		p = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return p, size
}
func buildMediaRecordList(items []*domain.Record, total int64, page, size int) contract.MediaRecordListResp {
	out := make([]contract.MediaRecordResp, len(items))
	for i := range items {
		out[i] = contract.MediaRecordResponse(items[i])
	}
	return contract.MediaRecordListResp{Items: out, Total: total, Page: page, Size: size}
}
func mediaRecordError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return response.ErrorWithMsg[any](c, response.NotFound, "影视记录不存在")
	case errors.Is(err, domain.ErrInvalidStatus), errors.Is(err, domain.ErrInvalidType):
		return response.ErrorWithMsg[any](c, response.ParamsError, "影视类型或观看状态无效")
	default:
		return err
	}
}
