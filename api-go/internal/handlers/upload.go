package handlers

import (
	"log"
	"net/http"

	m "github.com/cymoo/mint"
	e "github.com/cymoo/pebble/internal/errors"
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
		return nil, e.BadRequest()
	}

	defer file.Close()

	if header.Filename == "" {
		return nil, e.NotFound("invalid upload file name")
	}

	fileInfo, err := h.uploadService.UploadFile(header)
	if err != nil {
		log.Printf("error handling uploaded file: %v", err)
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
