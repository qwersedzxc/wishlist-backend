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
func validateImageContent(file io.Reader) (string, error) {
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
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return contentType, nil
		}
	}

	return "", fmt.Errorf("invalid file type: %s", contentType)
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
