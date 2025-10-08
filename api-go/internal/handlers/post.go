package handlers

import (
	"errors"
	"net/http"
	"time"

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

func (h *PostHandler) GetStats(r *http.Request) (*models.PostStats, error) {
	postCount, err := h.postService.GetCount(r.Context())
	if err != nil {
		return nil, err
	}

	tagCount, err := h.postService.GetCount(r.Context())
	if err != nil {
		return nil, err
	}

	dayCount, err := h.postService.GetActiveDays(r.Context())
	if err != nil {
		return nil, err
	}

	return &models.PostStats{
		PostCount: postCount,
		TagCount:  tagCount,
		DayCount:  dayCount,
	}, nil
}

func (h *PostHandler) GetDailyCounts(r *http.Request, query m.Query[models.DateRange]) ([]int64, error) {
	startDate, err := time.Parse(time.DateOnly, query.Value.StartDate)
	if err != nil {
		return nil, err
	}

	endDate, err := time.Parse(time.DateOnly, query.Value.EndDate)
	if err != nil {
		return nil, err
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	counts, err := h.postService.GetDailyCounts(r.Context(), startDate, endDate, query.Value.Offset*60)
	if err != nil {
		return nil, err
	}
	return counts, nil
}
