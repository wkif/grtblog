package handler

import (
	"context"
	"encoding/json"
	"strings"

	mediaapp "github.com/grtsinry43/grtblog-v2/server/internal/app/media"
)

func finalizeDraftText(ctx context.Context, mediaSvc *mediaapp.Service, value string) (string, error) {
	if mediaSvc == nil || value == "" {
		return value, nil
	}
	return mediaSvc.RewriteDraftURLs(ctx, value)
}

func finalizeDraftOptionalURL(ctx context.Context, mediaSvc *mediaapp.Service, value *string) (*string, error) {
	if mediaSvc == nil || value == nil || *value == "" {
		return value, nil
	}
	rewritten, err := mediaSvc.PromoteDraftURL(ctx, *value)
	if err != nil {
		return nil, err
	}
	return &rewritten, nil
}

func finalizeDraftURLs(ctx context.Context, mediaSvc *mediaapp.Service, values []string) ([]string, error) {
	if mediaSvc == nil || len(values) == 0 {
		return values, nil
	}
	rewritten := make([]string, len(values))
	for i, value := range values {
		next, err := mediaSvc.PromoteDraftURL(ctx, value)
		if err != nil {
			return nil, err
		}
		rewritten[i] = next
	}
	return rewritten, nil
}

func finalizeDraftExtInfo(ctx context.Context, mediaSvc *mediaapp.Service, extInfo []byte) ([]byte, error) {
	if mediaSvc == nil || len(extInfo) == 0 {
		return extInfo, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(extInfo, &payload); err != nil {
		return nil, err
	}
	if images, ok := payload["images"].([]any); ok {
		for _, item := range images {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			id, ok := obj["id"].(string)
			if !ok || strings.TrimSpace(id) == "" {
				continue
			}
			rewritten, err := mediaSvc.PromoteDraftURL(ctx, id)
			if err != nil {
				return nil, err
			}
			obj["id"] = rewritten
		}
	}
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return rewritten, nil
}
