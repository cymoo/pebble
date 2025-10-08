package handlers

import (
	"errors"
	"log"
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

func (h *PostHandler) GetPosts(r *http.Request, query m.Query[models.PostFilterOptions]) (*models.PostPagination, error) {
	posts, err := h.postService.Filter(r.Context(), query.Value, 10)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		return nil, err
	}

	// Determine the new cursor based on the last post's CreatedAt
	size := len(posts)
	cursor := int64(-1)
	if size > 0 {
		cursor = posts[size-1].CreatedAt
	}

	return &models.PostPagination{
		Posts:  posts,
		Cursor: cursor,
		Size:   int64(size),
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

func (h *PostHandler) CreatePost(r *http.Request, body m.JSON[models.CreatePostRequest]) (*models.CreateResponse, error) {
	rv, err := h.postService.Create(r.Context(), body.Value)
	if err != nil {
		return nil, err
	}
	return rv, nil
}

func (h *PostHandler) UpdatePost(r *http.Request, body m.JSON[models.UpdatePostRequest]) (m.StatusCode, error) {
	err := h.postService.Update(r.Context(), body.Value)
	if err != nil {
		return 0, err
	}
	return 204, nil
}

func (h *PostHandler) DeletePost(r *http.Request, payload m.JSON[models.DeletePostRequest]) (m.StatusCode, error) {
	if payload.Value.Hard {
		err := h.postService.HardDelete(r.Context(), payload.Value.ID)
		if err != nil {
			return 0, err
		}
	} else {
		err := h.postService.Delete(r.Context(), payload.Value.ID)
		if err != nil {
			log.Printf("Error deleting post: %v", err)
			return 0, err
		}
	}

	return 204, nil
}

func (h *PostHandler) RestorePost(r *http.Request, payload m.JSON[models.Id]) (m.StatusCode, error) {
	err := h.postService.Restore(r.Context(), payload.Value.Id)
	if err != nil {
		return 0, err
	}
	return 204, nil
}

func (h *PostHandler) ClearPosts(r *http.Request) (m.StatusCode, error) {
	ids, err := h.postService.ClearAll(r.Context())
	if err != nil {
		return 0, err
	}
	log.Printf("Cleared posts: %v", ids)
	return 204, nil
}
