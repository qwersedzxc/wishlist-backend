package v1

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const maxFileSize = 5 << 20 // 5MB

// UploadHandler обрабатывает загрузку файлов в S3-совместимое хранилище
type UploadHandler struct {
	log      *slog.Logger
	s3Client *s3.Client
	bucket   string
	baseURL  string
	enabled  bool
}

// S3Config конфигурация S3-совместимого хранилища
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	BaseURL         string
	Region          string
}

// Проверка реального типа файла (читаем первые 512 байт)
func validateImageContent(file io.ReadSeeker) (string, error) {
	// Читаем первые 512 байт для определения типа
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Определяем MIME тип по содержимому
	contentType := http.DetectContentType(buffer[:n])

	// Разрешенные типы
	allowedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
		"image/gif",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			// Перематываем файл обратно в начало
			_, err := file.Seek(0, io.SeekStart)
			if err != nil {
				return "", fmt.Errorf("failed to seek file: %w", err)
			}
			return contentType, nil
		}
	}

	return "", fmt.Errorf("invalid file type: %s (allowed: jpeg, png, webp, gif)", contentType)
}

func (c *S3Config) IsEmpty() bool {
	return c.Endpoint == "" || c.AccessKeyID == "" || c.SecretAccessKey == "" || c.Bucket == ""
}

func newUploadHandler(log *slog.Logger, cfg S3Config) *UploadHandler {
	if cfg.IsEmpty() {
		log.Warn("S3 config is empty or incomplete, upload handler will be disabled",
			"endpoint", cfg.Endpoint != "",
			"access_key", cfg.AccessKeyID != "",
			"secret_key", cfg.SecretAccessKey != "",
			"bucket", cfg.Bucket != "",
		)
		return &UploadHandler{
			log:     log,
			enabled: false,
		}
	}
	// Создаем кастомный resolver для Yandex Cloud
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Error("failed to load AWS config", "error", err)
		return &UploadHandler{
			log:     log,
			enabled: false,
		}
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = false // Yandex использует virtual-hosted style
	})

	log.Info("S3 client initialized",
		"endpoint", cfg.Endpoint,
		"bucket", cfg.Bucket,
		"region", cfg.Region)

	return &UploadHandler{
		log:      log,
		s3Client: s3Client,
		bucket:   cfg.Bucket,
		baseURL:  cfg.BaseURL,
		enabled:  true,
	}
}

// UploadResponse ответ при загрузке файла
type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		h.log.Error("S3 upload handler is disabled")
		render.Status(r, http.StatusServiceUnavailable)
		render.JSON(w, r, map[string]string{"error": "upload service is not available"})
		return
	}

	h.log.Info("upload image request received")
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		h.log.Error("failed to parse multipart form", "error", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "file too large or invalid form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.log.Error("failed to get form file", "error", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "no file provided"})
		return
	}
	defer file.Close()

	h.log.Info("file received", "filename", header.Filename, "size", header.Size, "content_type", header.Header.Get("Content-Type"))

	contentType, err := validateImageContent(file)
	if err != nil {
		h.log.Error("file validation failed", "error", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	ext := getFileExtension(header.Filename)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	h.log.Info("uploading to S3", "bucket", h.bucket, "key", filename)

	// Дополнительная проверка, что s3Client не nil, иначе паника
	if h.s3Client == nil {
		h.log.Error("S3 client is nil")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "upload service not properly initialized"})
		return
	}

	_, err = h.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(h.bucket),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         "public-read", // Делаем файл публично доступным
	})
	if err != nil {
		h.log.Error("failed to upload to S3",
			"error", err,
			"bucket", h.bucket,
			"key", filename)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": fmt.Sprintf("failed to upload file: %v", err)})
		return
	}

	fileURL := fmt.Sprintf("%s/%s", strings.TrimRight(h.baseURL, "/"), filename)
	h.log.Info("file uploaded to S3", "key", filename, "size", header.Size, "url", fileURL)

	render.JSON(w, r, UploadResponse{
		URL:      fileURL,
		Filename: filename,
		Size:     header.Size,
	})
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".jpg"
	}
	return strings.ToLower(ext)
}

// ProxyImage проксирует изображение из S3 для обхода CORS
func (h *UploadHandler) ProxyImage(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		http.Error(w, "upload service is not available", http.StatusServiceUnavailable)
		return
	}

	// Получаем путь к файлу из URL
	imagePath := r.URL.Query().Get("path")
	if imagePath == "" {
		http.Error(w, "missing path parameter", http.StatusBadRequest)
		return
	}

	// Извлекаем bucket и key из пути
	// Формат: wishlist-images/filename.jpg
	parts := strings.SplitN(imagePath, "/", 2)
	if len(parts) != 2 {
		h.log.Error("invalid image path format", "path", imagePath)
		http.Error(w, "invalid path format", http.StatusBadRequest)
		return
	}
	
	bucket := parts[0]
	key := parts[1]
	
	h.log.Info("proxying image from S3", "bucket", bucket, "key", key)

	// Загружаем изображение из S3 используя SDK
	result, err := h.s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		h.log.Error("failed to get object from S3", "error", err, "bucket", bucket, "key", key)
		http.Error(w, "failed to fetch image", http.StatusInternalServerError)
		return
	}
	defer result.Body.Close()

	// Читаем все данные в память сразу
	imageData, err := io.ReadAll(result.Body)
	if err != nil {
		h.log.Error("failed to read image data", "error", err)
		http.Error(w, "failed to read image", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки
	contentType := "image/jpeg" // fallback
	if result.ContentType != nil {
		contentType = *result.ContentType
		h.log.Info("S3 returned content type", "content_type", contentType)
	} else {
		h.log.Warn("S3 did not return content type, using fallback", "fallback", contentType)
	}
	
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(imageData)))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	// Отправляем все данные одним куском
	written, err := w.Write(imageData)
	if err != nil {
		h.log.Error("failed to write image data", "error", err)
		return
	}
	
	h.log.Info("image proxied successfully", "bytes", written, "content_type", contentType)
}
