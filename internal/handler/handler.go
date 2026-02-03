package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"taskS3/internal/config"
	"taskS3/internal/service"
)

type Handler struct {
	service service.ImageService
	log     *zap.Logger
	cfg     *config.Config
}

func NewHandler(service service.ImageService, cfg *config.Config, log *zap.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
		cfg:     cfg,
	}
}

func (h *Handler) UploadImage(c *gin.Context) {
	h.log.Info("Upload endpoint called")
	c.JSON(http.StatusOK, gin.H{"message": "Upload endpoint - working!"})
}

func (h *Handler) ProcessImages(c *gin.Context) {
	h.log.Info("Process endpoint called")
	c.JSON(http.StatusOK, gin.H{"message": "Process endpoint - working!"})
}

func (h *Handler) MoveImages(c *gin.Context) {
	h.log.Info("Move endpoint called")
	c.JSON(http.StatusOK, gin.H{"message": "Move endpoint - working!"})
}

func (h *Handler) ListImages(c *gin.Context) {
	h.log.Info("List endpoint called")
	c.JSON(http.StatusOK, gin.H{"images": []string{"image1.jpg", "image2.png"}})
}

func (h *Handler) HealthCheck(c *gin.Context) {
	h.log.Info("Health check called")
	c.JSON(http.StatusOK, gin.H{"status": "OK", "service": "TaskS3"})
}

func (h *Handler) GetUI(c *gin.Context) {
	h.log.Info("UI endpoint called")
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title": "TaskS3 Image Manager",
	})
}
