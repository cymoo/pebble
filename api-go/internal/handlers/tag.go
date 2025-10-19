package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	m "github.com/cymoo/mint"
	e "github.com/cymoo/pebble/internal/errors"
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
		log.Printf("error getting tags: %v", err)
		return nil, err
	}
	return tags, nil
}

func (h *TagHandler) RenameTag(r *http.Request, payload m.JSON[models.RenameTagRequest]) (m.StatusCode, error) {
	oldName := payload.Value.Name
	newName := payload.Value.NewName

	// Check for invalid hierarchy
	if strings.HasPrefix(newName, oldName+"/") {
		return 0, e.BadRequest(fmt.Sprintf("cannot move %q to a subtag of itself %q", oldName, newName))
	}

	err := h.tagService.RenameOrMerge(r.Context(), oldName, newName)
	if err != nil {
		log.Printf("error renaming tag %q to %q: %v", oldName, newName, err)
		return 0, err
	}
	return m.StatusCode(204), nil
}

func (h *TagHandler) DeleteTag(r *http.Request, payload m.JSON[models.Name]) (m.StatusCode, error) {
	tagName := payload.Value.Name
	err := h.tagService.DeleteAssociatedPosts(r.Context(), tagName)
	if err != nil {
		log.Printf("error delete tag %q: %v", tagName, err)
		return 0, err
	}
	return m.StatusCode(204), nil
}

func (h *TagHandler) StickTag(r *http.Request, payload m.JSON[models.StickyTagRequest]) (m.StatusCode, error) {
	tagName := payload.Value.Name
	err := h.tagService.InsertOrUpdate(r.Context(), tagName, payload.Value.Sticky)
	if err != nil {
		log.Printf("error updating tag %q: %v", tagName, err)
		return 0, err
	}
	return m.StatusCode(204), nil
}
