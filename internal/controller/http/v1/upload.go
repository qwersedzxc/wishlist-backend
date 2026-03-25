package v1

import (
	"context"
	"fmt"
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

func newUploadHandler(log *slog.Logger, cfg S3Config) *UploadHandler {
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
	}
}

// UploadResponse ответ при загрузке файла
type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
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

	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		h.log.Error("invalid file type", "content_type", contentType)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid file type. Only JPEG, PNG, WebP allowed"})
		return
	}

	ext := getFileExtension(header.Filename)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	h.log.Info("uploading to S3", "bucket", h.bucket, "key", filename)

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

func isValidImageType(contentType string) bool {
	for _, t := range []string{"image/jpeg", "image/jpg", "image/png", "image/webp"} {
		if contentType == t {
			return true
		}
	}
	return false
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".jpg"
	}
	return strings.ToLower(ext)
}
