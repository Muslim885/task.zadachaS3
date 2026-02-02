package service

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"taskS3/internal/config"
	"taskS3/internal/domain"
	"taskS3/internal/repository"
	"taskS3/pkg/utils"
)

type ImageService interface {
	UploadImage(ctx context.Context, fileBytes []byte, filename, contentType string) (*domain.Image, error)
	ProcessLocalImages(ctx context.Context) error
	ProcessAndMoveImages(ctx context.Context) error
	ListImages(ctx context.Context) ([]domain.Image, error)
}

type imageService struct {
	s3Repo repository.S3Repository
	cfg    *config.Config
	log    *zap.Logger
	proc   *utils.ImageProcessor
}

func NewImageService(s3Repo repository.S3Repository, cfg *config.Config, log *zap.Logger) ImageService {
	return &imageService{
		s3Repo: s3Repo,
		cfg:    cfg,
		log:    log,
		proc:   utils.NewImageProcessor(log),
	}
}

func (s *imageService) UploadImage(ctx context.Context, fileBytes []byte, filename, contentType string) (*domain.Image, error) {
	imageID := uuid.New().String()
	ext := filepath.Ext(filename)
	key := "images/" + imageID + ext

	reader := bytes.NewReader(fileBytes)

	if err := s.s3Repo.UploadFile(ctx, key, reader, int64(len(fileBytes)), contentType); err != nil {
		return nil, err
	}

	image := &domain.Image{
		ID:           imageID,
		OriginalName: filename,
		StoragePath:  key,
		Size:         int64(len(fileBytes)),
		ContentType:  contentType,
		UploadedAt:   time.Now(),
		Processed:    false,
	}

	s.log.Info("Image uploaded successfully",
		zap.String("id", imageID),
		zap.String("filename", filename),
		zap.Int64("size", image.Size))

	return image, nil
}

func (s *imageService) ProcessLocalImages(ctx context.Context) error {
	s.log.Info("Starting local image processing",
		zap.String("upload_dir", s.cfg.App.UploadDir))

	// Обработка и сжатие локальных изображений
	if err := s.proc.ProcessLocalImages(s.cfg.App.UploadDir, s.cfg.App.ProcessedDir, 80); err != nil {
		return err
	}

	// Загрузка обработанных изображений в S3
	entries, err := os.ReadDir(s.cfg.App.ProcessedDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(s.cfg.App.ProcessedDir, entry.Name())
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			s.log.Error("Failed to read processed file",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		contentType := "image/jpeg"
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
			contentType = "image/png"
		}

		if _, err := s.UploadImage(ctx, fileBytes, entry.Name(), contentType); err != nil {
			s.log.Error("Failed to upload processed image",
				zap.String("file", entry.Name()),
				zap.Error(err))
		}
	}

	return nil
}

func (s *imageService) ProcessAndMoveImages(ctx context.Context) error {
	s.log.Info("Starting image processing and moving")

	// Получение списка файлов из S3
	keys, err := s.s3Repo.ListFiles(ctx, "images/")
	if err != nil {
		return err
	}

	// Создание директории для перемещенных файлов
	movedDir := filepath.Join(s.cfg.App.ProcessedDir, "moved")
	if err := os.MkdirAll(movedDir, 0755); err != nil {
		return err
	}

	for _, key := range keys {
		// Скачивание файла
		reader, err := s.s3Repo.DownloadFile(ctx, key)
		if err != nil {
			s.log.Error("Failed to download file",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		// Сохранение локально
		filename := filepath.Base(key)
		destPath := filepath.Join(movedDir, filename)

		file, err := os.Create(destPath)
		if err != nil {
			reader.Close()
			s.log.Error("Failed to create local file",
				zap.String("path", destPath),
				zap.Error(err))
			continue
		}

		if _, err := io.Copy(file, reader); err != nil {
			reader.Close()
			file.Close()
			s.log.Error("Failed to copy file",
				zap.String("key", key),
				zap.String("dest", destPath),
				zap.Error(err))
			continue
		}

		reader.Close()
		file.Close()

		// Копирование в другую папку S3
		newKey := "processed/" + filename
		if err := s.s3Repo.CopyFile(ctx, key, newKey); err != nil {
			s.log.Error("Failed to copy file in S3",
				zap.String("source", key),
				zap.String("destination", newKey),
				zap.Error(err))
			continue
		}

		s.log.Info("Image processed and moved",
			zap.String("original_key", key),
			zap.String("new_key", newKey),
			zap.String("local_path", destPath))
	}

	return nil
}

func (s *imageService) ListImages(ctx context.Context) ([]domain.Image, error) {
	keys, err := s.s3Repo.ListFiles(ctx, "images/")
	if err != nil {
		return nil, err
	}

	var images []domain.Image
	for _, key := range keys {
		image := domain.Image{
			ID:           filepath.Base(key),
			StoragePath:  key,
			OriginalName: filepath.Base(key),
			UploadedAt:   time.Now(),
			ContentType:  "image/jpeg",
			Processed:    false,
		}

		// Определяем content type по расширению
		ext := filepath.Ext(key)
		if ext == ".png" {
			image.ContentType = "image/png"
		}

		images = append(images, image)
	}

	return images, nil
}
