package handler

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"taskS3/internal/service"
)

type Handler struct {
	service service.ImageService
	log     *zap.Logger
}

func NewHandler(service service.ImageService, log *zap.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		h.log.Error("Failed to get file from form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// Проверка размера файла
	if file.Size > 10*1024*1024 { // 10MB
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	// Проверка формата
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Only JPG, JPEG, PNG allowed"})
		return
	}

	fileBytes, err := file.Open()
	if err != nil {
		h.log.Error("Failed to open file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	defer fileBytes.Close()

	// Чтение файла
	buf := make([]byte, file.Size)
	_, err = fileBytes.Read(buf)
	if err != nil {
		h.log.Error("Failed to read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		if ext == ".png" {
			contentType = "image/png"
		} else {
			contentType = "image/jpeg"
		}
	}

	image, err := h.service.UploadImage(c.Request.Context(), buf, file.Filename, contentType)
	if err != nil {
		h.log.Error("Failed to upload image", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image uploaded successfully",
		"image":   image,
	})
}

func (h *Handler) ProcessImages(c *gin.Context) {
	if err := h.service.ProcessLocalImages(c.Request.Context()); err != nil {
		h.log.Error("Failed to process images", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Images processed successfully"})
}

func (h *Handler) MoveImages(c *gin.Context) {
	if err := h.service.ProcessAndMoveImages(c.Request.Context()); err != nil {
		h.log.Error("Failed to move images", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Images moved successfully"})
}

func (h *Handler) ListImages(c *gin.Context) {
	images, err := h.service.ListImages(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to list images", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"images": images})
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (h *Handler) GetUI(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}
