package handlers

import (
	"errors"
	"net/http"

	m "github.com/cymoo/mint"
	"github.com/cymoo/pebble/internal/models"
	"github.com/cymoo/pebble/internal/services"
)

type TagHandler struct {
	tagService *services.TagService
}

func NewTagHandler(tagService *services.TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

func (h *TagHandler) GetTags(r *http.Request) ([]models.TagWithPostCount, error) {
	tags, err := h.tagService.GetAllWithPostCount(r.Context())
	if err != nil {
		return nil, err
	}
	if tags == nil {
		return nil, errors.New("no tags found")
	}
	return tags, nil
}

func (h *TagHandler) RenameTag(r *http.Request, payload m.JSON[models.RenameTagRequest]) (m.StatusCode, error) {
	err := h.tagService.RenameOrMerge(r.Context(), payload.Value.Name, payload.Value.NewName)
	if err != nil {
		return 0, err
	}
	return m.StatusCode(204), nil
}

func (h *TagHandler) DeleteTag(r *http.Request, payload m.JSON[models.Name]) (m.StatusCode, error) {
	err := h.tagService.DeleteAssociatedPosts(r.Context(), payload.Value.Name)
	if err != nil {
		return 0, err
	}
	return m.StatusCode(204), nil
}

func (h *TagHandler) StickTag(r *http.Request, payload m.JSON[models.StickyTagRequest]) (m.StatusCode, error) {
	err := h.tagService.InsertOrUpdate(r.Context(), payload.Value.Name, payload.Value.Sticky)
	if err != nil {
		return 0, err
	}
	return m.StatusCode(204), nil
}
