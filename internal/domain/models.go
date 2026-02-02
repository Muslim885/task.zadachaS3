package domain

import (
	"time"
)

type Image struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	StoragePath  string    `json:"storage_path"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	UploadedAt   time.Time `json:"uploaded_at"`
	Processed    bool      `json:"processed"`
}
