package app

import (
	"net/http"
	"time"

	m "github.com/cymoo/mint"
	"github.com/cymoo/pebble/assets"
	e "github.com/cymoo/pebble/internal/errors"
	"github.com/cymoo/pebble/internal/handlers"
	"github.com/cymoo/pebble/internal/models"
	"github.com/cymoo/pebble/internal/services"
	"github.com/go-chi/chi/v5"
)

// NewApiRouter creates and returns a router for API endpoints
func NewApiRouter(app *App) *chi.Mux {
	r := chi.NewRouter()

	tagService := services.NewTagService(app.db)
	tagHandler := handlers.NewTagHandler(tagService)

	postService := services.NewPostService(app.db)
	postHandler := handlers.NewPostHandler(postService, tagService, app.fts)

	uploadService := services.NewUploadService(&app.config.Upload)
	uploadHandler := handlers.NewUploadHandler(uploadService)

	authService := services.NewAuthService()

	// Use simple auth check middleware for all routes except /api/login
	r.Use(SimpleAuthCheck(authService, "/api/login"))

	// handleLogin processes login requests by validating the provided password
	handleLogin := func(payload m.JSON[models.LoginRequest]) (m.StatusCode, error) {
		if authService.IsValidToken(payload.Value.Password) {
			return http.StatusNoContent, nil
		} else {
			return 0, e.Unauthorized("password is wrong")
		}
	}

	// Use rate limiting middleware for login route
	r.With(RateLimit(app.redis, 60*time.Second, 5)).Post("/login", m.H(handleLogin))

	// A simple endpoint to verify authentication
	// Nginx can use this to check if the token is valid, and handle uploads accordingly
	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	r.Get("/hello", m.H(postHandler.HelloWorld))

	r.Get("/get-tags", m.H(tagHandler.GetTags))
	r.Post("/rename-tag", m.H(tagHandler.RenameTag))
	r.Post("/delete-tag", m.H(tagHandler.DeleteTag))
	r.Post("/stick-tag", m.H(tagHandler.StickTag))

	r.Get("/search", m.H(postHandler.SearchPosts))
	r.Get("/get-posts", m.H(postHandler.GetPosts))
	r.Get("/get-post", m.H(postHandler.GetPost))
	r.Post("/create-post", m.H(postHandler.CreatePost))
	r.Post("/update-post", m.H(postHandler.UpdatePost))
	r.Post("/delete-post", m.H(postHandler.DeletePost))
	r.Post("/restore-post", m.H(postHandler.RestorePost))
	r.Post("/clear-posts", m.H(postHandler.ClearPosts))

	r.Get("/get-overall-counts", m.H(postHandler.GetStats))
	r.Get("/get-daily-post-counts", m.H(postHandler.GetDailyCounts))

	r.Post("/upload", m.H(uploadHandler.UploadFile))
	r.Get("/upload", m.H(uploadHandler.SimpleFileForm))

	return r
}

// NewPageRouter creates and returns a router for page endpoints
func NewPageRouter(app *App) *chi.Mux {
	r := chi.NewRouter()

	pageHandler, err := handlers.NewPostPageHandler(app.db, assets.TemplateFS())
	if err != nil {
		panic("failed to create page handler: " + err.Error())
	}

	r.Get("/", pageHandler.PostList)
	r.Get("/{id}", pageHandler.PostItem)

	return r
}
