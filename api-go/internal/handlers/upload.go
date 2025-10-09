package handlers

import (
	"net/http"

	m "github.com/cymoo/mint"
	"github.com/cymoo/pebble/internal/models"
	"github.com/cymoo/pebble/internal/services"
)

type UploadHandler struct {
	uploadService *services.UploadService
}

func NewUploadHandler(uploadService *services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

func (h *UploadHandler) UploadFile(r *http.Request) (*models.FileInfo, error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		// return
		return nil, err
	}
	defer file.Close()

	fileInfo, err := h.uploadService.UploadFile(header)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		// return
		return nil, err
	}
	return fileInfo, nil
}

func (h *UploadHandler) SimpleFileForm() m.HTML {
	return `<!doctype html>
  <html>
    <head>
      <title>Upload file</title>
      <meta charset="utf-8">
    </head>
    <body>
      <form action="" method="post" enctype="multipart/form-data">
        <input type="file" name="file" multiple>
          <button type="submit">Upload</button>
        </form>
    </body>
  </html>`
}
