package services

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cymoo/pebble/internal/config"
	"github.com/cymoo/pebble/internal/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/webp"
)

// // UploadConfig 文件上传配置
// type UploadConfig struct {
// 	UploadDir     string
// 	UploadURL     string
// 	ThumbnailSize int
// 	ImageFormats  map[string]bool
// }

// // FileInfo 文件信息
// type FileInfo struct {
// 	URL      string `json:"url"`
// 	Size     string `json:"size,omitempty"`
// 	ThumbURL string `json:"thumb_url,omitempty"`
// 	Width    int    `json:"width,omitempty"`
// 	Height   int    `json:"height,omitempty"`
// }

// UploadService 文件上传服务
type UploadService struct {
	config *config.UploadConfig
}

// NewUploadService 创建新的文件上传服务
func NewUploadService(config *config.UploadConfig) *UploadService {
	if config.ThumbSize == 0 {
		config.ThumbSize = 200
	}

	// 确保上传目录存在
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		panic(fmt.Sprintf("无法创建上传目录: %v", err))
	}

	return &UploadService{
		config: config,
	}
}

// UploadFile 上传文件
func (s *UploadService) UploadFile(fileHeader *multipart.FileHeader) (*models.FileInfo, error) {
	// 验证文件名
	if fileHeader.Filename == "" {
		return nil, errors.New("无效的文件名")
	}

	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("无法打开上传的文件: %w", err)
	}
	defer file.Close()

	// 生成安全文件名
	secureFileName := s.generateSecureFilename(fileHeader.Filename)
	filePath := filepath.Join(s.config.BasePath, secureFileName)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法创建文件: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		// 如果复制失败，删除已创建的文件
		os.Remove(filePath)
		return nil, fmt.Errorf("无法保存文件: %w", err)
	}

	// 获取文件类型
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		// 尝试从文件内容检测类型
		if detectedType, err := s.detectContentType(filePath); err == nil {
			contentType = detectedType
		}
	}

	// 处理文件
	if s.isImage(contentType) {
		return s.processImageFile(filePath, contentType)
	}
	return s.processRegularFile(filePath)
}

// processRegularFile 处理普通文件
func (s *UploadService) processRegularFile(filePath string) (*models.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法获取文件信息: %w", err)
	}

	fileName := filepath.Base(filePath)
	size := uint64(fileInfo.Size())

	return &models.FileInfo{
		URL:  s.buildFileURL(fileName),
		Size: &size,
	}, nil
}

// processImageFile 处理图片文件
func (s *UploadService) processImageFile(filePath, contentType string) (*models.FileInfo, error) {
	// 读取图片
	img, err := s.decodeImage(filePath, contentType)
	if err != nil {
		return nil, fmt.Errorf("无法解码图片: %w", err)
	}

	// 处理EXIF旋转
	if s.needsExifRotation(contentType) {
		if err := s.handleExifRotation(filePath, img); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("处理EXIF旋转失败: %v\n", err)
		}
	}

	// 生成缩略图
	thumbURL, err := s.generateThumbnail(filePath, img)
	if err != nil {
		return nil, fmt.Errorf("无法生成缩略图: %w", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法获取文件信息: %w", err)
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

// decodeImage 解码图片
func (s *UploadService) decodeImage(filePath, contentType string) (image.Image, error) {
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

// handleExifRotation 处理EXIF旋转
func (s *UploadService) handleExifRotation(filePath string, img image.Image) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	exifData, err := exif.Decode(file)
	if err != nil {
		// 没有EXIF数据是正常情况
		return nil
	}

	orientation, err := exifData.Get(exif.Orientation)
	if err != nil {
		// 没有方向信息
		return nil
	}

	orientationVal, err := orientation.Int(0)
	if err != nil {
		return err
	}

	var rotated image.Image
	switch orientationVal {
	case 6, 8: // 需要旋转90度或270度
		rotated = imaging.Rotate90(img)
	case 3: // 需要旋转180度
		rotated = imaging.Rotate180(img)
	default:
		// 不需要旋转
		return nil
	}

	// 保存旋转后的图片
	return s.saveImage(filePath, rotated)
}

// generateThumbnail 生成缩略图
func (s *UploadService) generateThumbnail(originalPath string, img image.Image) (string, error) {
	fileName := filepath.Base(originalPath)
	thumbFileName := "thumb_" + fileName
	thumbPath := filepath.Join(s.config.BasePath, thumbFileName)

	// 生成缩略图
	// TODO: 高度自适应
	thumbnail := imaging.Thumbnail(img, s.config.ThumbSize, s.config.ThumbSize, imaging.Lanczos)

	if err := s.saveImage(thumbPath, thumbnail); err != nil {
		return "", err
	}

	return s.buildFileURL(thumbFileName), nil
}

// saveImage 保存图片
func (s *UploadService) saveImage(filePath string, img image.Image) error {
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

// generateSecureFilename 生成安全文件名
func (s *UploadService) generateSecureFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)

	// 清理文件名，只保留字母数字、连字符和下划线
	var cleanName strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			cleanName.WriteRune(r)
		} else {
			cleanName.WriteRune('_')
		}
	}

	// 添加时间戳和UUID防止冲突
	timestamp := time.Now().Format("20060102150405")
	uniqueID := uuid.New().String()[:8]

	return fmt.Sprintf("%s_%s_%s%s", cleanName.String(), timestamp, uniqueID, ext)
}

// buildFileURL 构建文件URL
func (s *UploadService) buildFileURL(fileName string) string {
	return s.config.BaseURL + "/" + fileName
}

// isImage 检查是否为图片
func (s *UploadService) isImage(contentType string) bool {
	if !strings.HasPrefix(contentType, "image/") {
		return false
	}

	format := strings.ToLower(strings.TrimPrefix(contentType, "image/"))
	return Contains(s.config.ImageFormats, format)

}

// needsExifRotation 检查是否需要EXIF旋转
func (s *UploadService) needsExifRotation(contentType string) bool {
	format := strings.ToLower(strings.TrimPrefix(contentType, "image/"))
	switch format {
	case "jpeg", "jpg":
		return true
	default:
		return false
	}
}

// detectContentType 检测文件内容类型
func (s *UploadService) detectContentType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取前512字节用于类型检测
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}

// Cleanup 清理服务
func (s *UploadService) Cleanup() error {
	// 这里可以添加清理逻辑，如关闭连接等
	return nil
}

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
