package server

import (
	"context"
	"fmt"
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

func New(cfg *config.Config, log *zap.Logger) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.LoadHTMLGlob("web/templates/*")

	s3Repo, err := repository.NewS3Repository(&cfg.S3, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 repository: %w", err)
	}

	imageService := service.NewImageService(s3Repo, cfg, log)

	h := handler.NewHandler(imageService, cfg, log)

	router.GET("/", h.GetUI)
	router.GET("/health", h.HealthCheck)

	api := router.Group("/api")
	{
		api.POST("/upload", h.UploadImage)
		api.POST("/process", h.ProcessImages)
		api.POST("/move", h.MoveImages)
		api.GET("/images", h.ListImages)
	}

	router.Static("/static", "./web/static")
	router.Static("/uploads", "./uploads")

	server := &Server{
		httpServer: &http.Server{
			Addr:           cfg.Server.Host + ":" + cfg.Server.Port,
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1 MB
		},
		cfg: cfg,
		log: log,
	}

	log.Info("Server created successfully",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port))

	return server, nil
}

func (s *Server) Run() error {
	s.log.Info("Server is running",
		zap.String("host", s.cfg.Server.Host),
		zap.String("port", s.cfg.Server.Port),
		zap.String("address", s.httpServer.Addr))

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down server")
	return s.httpServer.Shutdown(ctx)
}
