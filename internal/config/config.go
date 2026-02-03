package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	S3     S3Config
	App    AppConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Region          string
}

type AppConfig struct {
	UploadDir      string
	ProcessedDir   string
	MaxUploadSize  int64
	AllowedFormats []string
}

func Load() (*Config, error) {
	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("S3_ENDPOINT", "localhost:9000")
	viper.SetDefault("S3_ACCESS_KEY_ID", "minioadmin")
	viper.SetDefault("S3_SECRET_ACCESS_KEY", "minioadmin")
	viper.SetDefault("S3_USE_SSL", false)
	viper.SetDefault("S3_BUCKET_NAME", "images")
	viper.SetDefault("S3_REGION", "us-east-1")
	viper.SetDefault("APP_UPLOAD_DIR", "./uploads/images")
	viper.SetDefault("APP_PROCESSED_DIR", "./uploads/processed")
	viper.SetDefault("APP_MAX_UPLOAD_SIZE", 10*1024*1024) // 10MB
	viper.SetDefault("APP_ALLOWED_FORMATS", []string{".jpg", ".jpeg", ".png"})

	viper.AutomaticEnv()

	cfg := &Config{
		Server: ServerConfig{
			Host: viper.GetString("SERVER_HOST"),
			Port: viper.GetString("SERVER_PORT"),
		},
		S3: S3Config{
			Endpoint:        viper.GetString("S3_ENDPOINT"),
			AccessKeyID:     viper.GetString("S3_ACCESS_KEY_ID"),
			SecretAccessKey: viper.GetString("S3_SECRET_ACCESS_KEY"),
			UseSSL:          viper.GetBool("S3_USE_SSL"),
			BucketName:      viper.GetString("S3_BUCKET_NAME"),
			Region:          viper.GetString("S3_REGION"),
		},
		App: AppConfig{
			UploadDir:      viper.GetString("APP_UPLOAD_DIR"),
			ProcessedDir:   viper.GetString("APP_PROCESSED_DIR"),
			MaxUploadSize:  viper.GetInt64("APP_MAX_UPLOAD_SIZE"),
			AllowedFormats: viper.GetStringSlice("APP_ALLOWED_FORMATS"),
		},
	}

	if err := createDirs(cfg); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return cfg, nil
}

func createDirs(cfg *Config) error {
	dirs := []string{
		cfg.App.UploadDir,
		cfg.App.ProcessedDir,
		filepath.Join(cfg.App.ProcessedDir, "moved"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
