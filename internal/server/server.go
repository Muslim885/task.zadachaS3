package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"taskS3/internal/config"
	"taskS3/internal/handler"
	"taskS3/internal/repository"
	"taskS3/internal/service"
)

type Server struct {
	httpServer *http.Server
	cfg        *config.Config
	log        *zap.Logger
}

func New(cfg *config.Config, log *zap.Logger) *Server {
	router := gin.Default()

	// Настройка шаблонов
	router.LoadHTMLGlob("web/templates/*")

	// Инициализация репозитория
	s3Repo, err := repository.NewS3Repository(&cfg.S3, log)
	if err != nil {
		log.Fatal("Failed to create S3 repository", zap.Error(err))
	}

	// Инициализация сервиса
	imageService := service.NewImageService(s3Repo, cfg, log)

	// Инициализация хендлера
	h := handler.NewHandler(imageService, log)

	// Настройка маршрутов
	router.GET("/", h.GetUI)
	router.GET("/health", h.HealthCheck)
	router.POST("/api/upload", h.UploadImage)
	router.POST("/api/process", h.ProcessImages)
	router.POST("/api/move", h.MoveImages)
	router.GET("/api/images", h.ListImages)

	// Статические файлы
	router.Static("/static", "./web/static")
	router.Static("/uploads", "./uploads")

	return &Server{
		httpServer: &http.Server{
			Addr:           cfg.Server.Host + ":" + cfg.Server.Port,
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		cfg: cfg,
		log: log,
	}
}

func (s *Server) Run() error {
	s.log.Info("Starting server",
		zap.String("host", s.cfg.Server.Host),
		zap.String("port", s.cfg.Server.Port),
		zap.String("address", s.httpServer.Addr))

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down server")
	return s.httpServer.Shutdown(ctx)
}
