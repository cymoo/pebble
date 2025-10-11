package handlers

import (
	"context"
	"log"
	"net/http"
	"sort"
	"time"

	m "github.com/cymoo/mint"
	e "github.com/cymoo/pebble/internal/errors"
	"github.com/cymoo/pebble/internal/models"
	"github.com/cymoo/pebble/internal/services"

	"github.com/cymoo/pebble/pkg/fulltext"
	"github.com/cymoo/pebble/pkg/util"
)

type PostHandler struct {
	postService *services.PostService
	tagService  *services.TagService
	fts         *fulltext.FullTextSearch
}

func NewPostHandler(postService *services.PostService, tagService *services.TagService, fts *fulltext.FullTextSearch) *PostHandler {
	return &PostHandler{postService: postService, tagService: tagService, fts: fts}
}

func (h *PostHandler) HelloWorld() string {
	return "hello world"
}

func (h *PostHandler) SearchPosts(r *http.Request, query m.Query[models.SearchRequest]) (*models.PostPagination, error) {
	ctx := r.Context()

	tokens, results, err := h.fts.Search(ctx, query.Value.Query)
	if err != nil {
		log.Printf("error searching: %v", err)
		return nil, e.InternalError()
	}

	if len(results) == 0 {
		return &models.PostPagination{
			Posts:  []models.Post{},
			Cursor: -1,
			Size:   0,
		}, nil
	}

	// build a map from ID to Score
	idToScore := make(map[int64]float64, len(results))
	ids := make([]int64, 0, len(results))

	for _, result := range results {
		idToScore[result.ID] = result.Score
		ids = append(ids, result.ID)
	}

	// get posts by IDs
	posts, err := h.postService.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// process each post's content and score
	for i := range posts {
		score, exists := idToScore[posts[i].ID]
		if exists {
			// highlight all occurrences of tokens in the content
			posts[i].Content = util.Highlight(posts[i].Content, tokens)
			posts[i].Score = &score
		}
	}

	// order by score desc
	sort.Slice(posts, func(i, j int) bool {
		scoreI, existsI := idToScore[posts[i].ID]
		scoreJ, existsJ := idToScore[posts[j].ID]

		if !existsI && !existsJ {
			return false
		}
		if !existsI {
			return false
		}
		if !existsJ {
			return true
		}
		return scoreI > scoreJ
	})

	size := int64(len(posts))

	return &models.PostPagination{
		Posts:  posts,
		Cursor: -1,
		Size:   size,
	}, nil
}

func (h *PostHandler) GetPosts(r *http.Request, query m.Query[models.FilterPostRequest]) (*models.PostPagination, error) {
	posts, err := h.postService.Filter(r.Context(), query.Value, 10)
	if err != nil {
		log.Printf("error getting posts: %v", err)
		return nil, e.InternalError()
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
		log.Printf("error getting post: %v", err)
		return nil, e.InternalError()
	}

	if post == nil {
		return nil, e.NotFound("post not found")
	}
	return post, nil
}

func (h *PostHandler) GetStats(r *http.Request) (*models.PostStats, error) {
	postCount, err := h.postService.GetCount(r.Context())
	if err != nil {
		return nil, err
	}

	tagCount, err := h.tagService.GetCount(r.Context())
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
		return nil, e.BadRequest("end_date must be after start_date")
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

	go func() {
		ctx := context.Background()
		if err := h.fts.Index(ctx, rv.ID, body.Value.Content); err != nil {
			log.Printf("error indexing post %d: %v", rv.ID, err)
		}
	}()

	return rv, nil
}

func (h *PostHandler) UpdatePost(r *http.Request, body m.JSON[models.UpdatePostRequest]) (m.StatusCode, error) {
	err := h.postService.Update(r.Context(), body.Value)
	if err != nil {
		return 0, err
	}

	if body.Value.Content != nil {
		go func() {
			ctx := context.Background()
			if err := h.fts.Reindex(ctx, body.Value.ID, *body.Value.Content); err != nil {
				log.Printf("error reindexing post %d: %v", body.Value.ID, err)
			}
		}()
	}

	return 204, nil
}

func (h *PostHandler) DeletePost(r *http.Request, payload m.JSON[models.DeletePostRequest]) (m.StatusCode, error) {
	if payload.Value.Hard {
		err := h.postService.HardDelete(r.Context(), payload.Value.ID)
		if err != nil {
			return 0, err
		}

		go func() {
			ctx := context.Background()
			if err := h.fts.Deindex(ctx, payload.Value.ID); err != nil {
				log.Printf("error deleting post %d from index: %v", payload.Value.ID, err)
			}
		}()

	} else {
		err := h.postService.Delete(r.Context(), payload.Value.ID)
		if err != nil {
			log.Printf("error deleting post: %v", err)
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
	log.Printf("cleared posts: %v", ids)

	go func() {
		ctx := context.Background()
		for _, id := range ids {
			if err := h.fts.Deindex(ctx, id); err != nil {
				log.Printf("error deleting post %d from index: %v", id, err)
			}
		}
	}()

	return 204, nil
}
