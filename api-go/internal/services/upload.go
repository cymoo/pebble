package services

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/cymoo/pebble/internal/config"
	"github.com/cymoo/pebble/internal/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/webp"
)

var invalidCharsRegex = regexp.MustCompile(`[^\w\-.\p{Han}]+`)

type UploadService struct {
	config *config.UploadConfig
}

func NewUploadService(config *config.UploadConfig) *UploadService {
	if config.ThumbWidth == 0 {
		config.ThumbWidth = 200
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}

	return &UploadService{
		config: config,
	}
}

// UploadFile handles the file upload process
// It saves the file, processes images, and returns FileInfo
func (s *UploadService) UploadFile(fileHeader *multipart.FileHeader) (*models.FileInfo, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	secureFileName := generateSecureFilename(fileHeader.Filename, 8)
	filePath := filepath.Join(s.config.BasePath, secureFileName)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// Copy the file content
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}
	dst.Close()

	// Get content type from header
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		// detectContentType reads the first 512 bytes of the file to determine its content type.
		if detectedType, err := detectContentType(filePath); err == nil {
			contentType = detectedType
		}
	}

	if s.isImage(contentType) {
		return s.processImageFile(filePath, contentType)
	}
	return s.processRegularFile(filePath)
}

// processRegularFile handles non-image files
// It simply returns the FileInfo with URL and size
func (s *UploadService) processRegularFile(filePath string) (*models.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileName := filepath.Base(filePath)
	size := uint64(fileInfo.Size())

	return &models.FileInfo{
		URL:  s.buildFileURL(fileName),
		Size: &size,
	}, nil
}

// processImageFile handles image-specific processing like EXIF rotation and thumbnail generation
// It returns the FileInfo with URL, thumbnail URL, size, width, and height
func (s *UploadService) processImageFile(filePath, contentType string) (*models.FileInfo, error) {
	// Read the image
	img, err := decodeImage(filePath, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Handle EXIF rotation
	if needsExifRotation(contentType) {
		img, err = handleExifRotation(filePath, img)
		if err != nil {
			// Log errors, but do not fail the upload
			log.Printf("failed to handle EXIF rotation: %v", err)
		}
	}

	// Handle thumbnail generation
	thumbURL, err := s.generateThumbnail(filePath, img)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileName := filepath.Base(filePath)
	size := uint64(fileInfo.Size())
	bounds := img.Bounds()
	width := uint32(bounds.Dx())
	height := uint32(bounds.Dy())

	return &models.FileInfo{
		URL:      s.buildFileURL(fileName),
		ThumbURL: &thumbURL,
		Size:     &size,
		Width:    &width,
		Height:   &height,
	}, nil
}

// generates a thumbnail with a fixed width, maintaining aspect ratio
// It returns the thumbnail URL
func (s *UploadService) generateThumbnail(originalPath string, img image.Image) (string, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= int(s.config.ThumbWidth) {
		return s.buildFileURL(filepath.Base(originalPath)), nil
	}

	thumbHeight := int(int64(height) * int64(s.config.ThumbWidth) / int64(width))
	thumbnail := imaging.Thumbnail(img, int(s.config.ThumbWidth), thumbHeight, imaging.Lanczos)

	fileName := filepath.Base(originalPath)
	thumbFileName := "thumb_" + fileName
	thumbPath := filepath.Join(s.config.BasePath, thumbFileName)

	if err := saveImage(thumbPath, thumbnail); err != nil {
		return "", err
	}

	return s.buildFileURL(thumbFileName), nil
}

// buildFileURL constructs the file URL
func (s *UploadService) buildFileURL(fileName string) string {
	return s.config.BaseURL + "/" + fileName
}

// isImage checks if the content type represents an image
func (s *UploadService) isImage(contentType string) bool {
	if !strings.HasPrefix(contentType, "image/") {
		return false
	}

	format := strings.ToLower(strings.TrimPrefix(contentType, "image/"))
	return slices.Contains(s.config.ImageFormats, format)

}

// needsExifRotation only applies to JPEG images
func needsExifRotation(contentType string) bool {
	format := strings.ToLower(strings.TrimPrefix(contentType, "image/"))
	switch format {
	case "jpeg", "jpg":
		return true
	default:
		return false
	}
}

// handleExifRotation handles only JPEG images as EXIF is primarily used in JPEGs
func handleExifRotation(filePath string, img image.Image) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return img, err
	}
	defer file.Close()

	exifData, err := exif.Decode(file)
	if err != nil {
		return img, nil
	}

	orientation, err := exifData.Get(exif.Orientation)
	if err != nil {
		return img, nil
	}

	orientationVal, err := orientation.Int(0)
	if err != nil {
		return img, err
	}

	var rotated image.Image
	switch orientationVal {
	case 3:
		rotated = imaging.Rotate180(img)
	case 6:
		rotated = imaging.Rotate270(img)
	case 8:
		rotated = imaging.Rotate90(img)
	default:
		return img, nil
	}

	if err := saveImage(filePath, rotated); err != nil {
		return img, err
	}
	return rotated, nil
}

// decodeImage decodes JPEG, PNG, WEBP and other formats supported by the imaging package
func decodeImage(filePath, contentType string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	switch {
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		return jpeg.Decode(file)
	case strings.Contains(contentType, "png"):
		return png.Decode(file)
	case strings.Contains(contentType, "webp"):
		return webp.Decode(file)
	default:
		return imaging.Decode(file)
	}
}

// saveImage saves the image in the appropriate format based on the file extension
func saveImage(filePath string, img image.Image) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	case ".png":
		return png.Encode(file, img)
	default:
		return imaging.Encode(file, img, imaging.JPEG)
	}
}

// detectContentType detects the content type of a file by reading its first 512 bytes
func detectContentType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read 512 bytes as per http.DetectContentType documentation
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}

// generateSecureFilename generates a secure filename with UUID suffix
// It preserves Chinese characters and common filename characters
// uuidLength should be between 8 and 32
func generateSecureFilename(filename string, uuidLength int) string {
	if uuidLength < 8 || uuidLength > 32 {
		uuidLength = 8
	}

	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "file"
	}

	// Sanitize filename: keep word characters, hyphens, dots, and Chinese characters
	sanitized := invalidCharsRegex.ReplaceAllString(filename, "_")

	// Split name and extension
	ext := filepath.Ext(sanitized)
	name := strings.TrimSuffix(sanitized, ext)

	// Ensure name is not empty after sanitization
	if name == "" || name == "_" {
		name = "file"
	}

	// Generate UUID suffix
	uuidStr := strings.ReplaceAll(uuid.New().String(), "-", "")
	if len(uuidStr) > uuidLength {
		uuidStr = uuidStr[:uuidLength]
	}

	// Build final filename: name.uuid.ext or name.uuid (if no extension)
	if ext != "" {
		return fmt.Sprintf("%s.%s%s", name, uuidStr, ext)
	}
	return fmt.Sprintf("%s.%s", name, uuidStr)
}
