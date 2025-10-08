package handlers

import (
	"errors"
	"net/http"

	m "github.com/cymoo/mint"
	"github.com/cymoo/pebble/internal/models"
	"github.com/cymoo/pebble/internal/services"
)

type PostHandler struct {
	postService *services.PostService
}

func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) HelloWorld() string {
	return "hello world"
}

func (h *PostHandler) GetPosts(r *http.Request, query m.Query[models.PostFilterOptions]) (models.PostPagination, error) {
	posts, _ := h.postService.Filter(r.Context(), query.Value, 10)
	return models.PostPagination{
		Posts:  posts,
		Cursor: 0,
		Size:   int64(len(posts)),
	}, nil
}

func (h *PostHandler) GetPost(r *http.Request, query m.Query[models.Id]) (*models.Post, error) {
	post, err := h.postService.FindByID(r.Context(), query.Value.Id)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, errors.New("post not found")
	}
	return post, nil
}
