package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"

	return config.Build()
}

func NewSugared() (*zap.SugaredLogger, error) {
	logger, err := New()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
