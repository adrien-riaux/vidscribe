package main

import (
	"log"
	"os"
)

func run(cfg Config) error {
	log.Println("Fetching uploads playlist...")
	playlistID, err := getUploadsPlaylistID(cfg.YouTubeAPIKey, cfg.ChannelID)
	if err != nil {
		return err
	}

	log.Println("Fetching latest video...")
	video, err := getLatestVideo(cfg.YouTubeAPIKey, playlistID)
	if err != nil {
		return err
	}
	log.Printf("Latest: [%s] %s (%s)", video.ID, video.Title, video.PublishedAt)

	lastID := loadLastVideoID(cfg)
	if lastID == video.ID {
		log.Println("No new video since last run — nothing to do.")
		return nil
	}

	log.Println("New video found! Asking Gemini for a deep summary...")
	summary, err := summarizeVideo(cfg.GeminiAPIKey, video)
	if err != nil {
		return err
	}

	// Save to local markdown file (useful locally; on Cloud Run logs capture output)
	if err := saveToFile(video, summary); err != nil {
		log.Printf("saveToFile: %v", err)
	}

	// Send email notification
	if err := sendEmail(cfg, video, summary); err != nil {
		log.Printf("sendEmail: %v", err)
	}

	// Persist state so we don't re-summarize next run
	if err := saveLastVideoID(cfg, video.ID); err != nil {
		log.Printf("saveLastVideoID: %v", err)
	}

	log.Println("Done!")
	return nil
}

func main() {
	cfg := loadConfig()

	if err := run(cfg); err != nil {
		log.Fatalf("Fatal error: %v", err)
		os.Exit(1)
	}
}
