package handlers

import (
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

func (h *PostHandler) UpdatePost(post m.JSON[models.UpdatePostRequest]) m.StatusCode {
	return 204
}
