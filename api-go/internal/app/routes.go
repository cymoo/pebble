package app

import (
	m "github.com/cymoo/mint"
	"github.com/cymoo/pebble/internal/handlers"
	"github.com/cymoo/pebble/internal/services"
	"github.com/go-chi/chi"
)

func NewApiRouter(app *App) *chi.Mux {
	r := chi.NewRouter()

	tagService := services.NewTagService(app.db)
	tagHandler := handlers.NewTagHandler(tagService)

	postService := services.NewPostService(app.db)
	postHandler := handlers.NewPostHandler(postService, tagService, app.fts)

	uploadService := services.NewUploadService(&app.config.Upload)
	uploadHandler := handlers.NewUploadHandler(uploadService)

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
