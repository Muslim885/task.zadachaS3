package utils

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

type ImageProcessor struct {
	log *zap.Logger
}

func NewImageProcessor(log *zap.Logger) *ImageProcessor {
	return &ImageProcessor{log: log}
}

func (p *ImageProcessor) CompressImage(inputPath string, quality int) ([]byte, string, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	contentType := "image/jpeg"

	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, "", err
	}

	p.log.Info("Image compressed",
		zap.String("input", inputPath),
		zap.Int("quality", quality),
		zap.Int("size", buf.Len()))

	return buf.Bytes(), contentType, nil
}

func (p *ImageProcessor) ProcessLocalImages(inputDir, outputDir string, quality int) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		inputPath := filepath.Join(inputDir, entry.Name())
		ext := strings.ToLower(filepath.Ext(entry.Name()))

		if ext != ".jpg" && ext != ".jpeg" {
			p.log.Warn("Skipping unsupported file", zap.String("file", entry.Name()))
			continue
		}

		data, _, err := p.CompressImage(inputPath, quality)
		if err != nil {
			p.log.Error("Failed to compress image",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		outputPath := filepath.Join(outputDir, entry.Name())
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			p.log.Error("Failed to save compressed image",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		p.log.Info("Image processed",
			zap.String("input", inputPath),
			zap.String("output", outputPath))
	}

	return nil
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
