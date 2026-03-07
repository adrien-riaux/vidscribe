package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

const stateKey = "last_video_id.txt"

func loadLastVideoID(cfg Config) string {
	if cfg.StateBucket != "" {
		id, err := readFromGCS(cfg.StateBucket, stateKey)
		if err != nil {
			// Object likely doesn't exist yet on first run — that's fine
			log.Printf("GCS state not found (first run?): %v", err)
			return ""
		}
		return id
	}
	// Local file fallback
	data, err := os.ReadFile(cfg.StateFile)
	if err != nil {
		return ""
	}

	return string(bytes.TrimSpace(data))
}

func saveLastVideoID(cfg Config, videoID string) error {
	if cfg.StateBucket != "" {
		return writeToGCS(cfg.StateBucket, stateKey, videoID)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.StateFile), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(cfg.StateFile, []byte(videoID), 0644)
}

func readFromGCS(bucket, object string) (string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("gcs client: %w", err)
	}
	defer client.Close()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(data)), nil
}

func writeToGCS(bucket, object, content string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("gcs client: %w", err)
	}
	defer client.Close()

	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err := wc.Write([]byte(content)); err != nil {
		return fmt.Errorf("gcs write: %w", err)
	}

	return wc.Close()
}
