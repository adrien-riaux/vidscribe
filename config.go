package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// API keys
	YouTubeAPIKey string
	GeminiAPIKey  string

	// Target Channel
	ChannelID string

	// State with GCS in prod and local fallback
	StateBucket string
	StateFile   string

	// Email
	EmailFrom    string
	EmailTo      string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
}

func loadConfig() Config {
	_ = godotenv.Load(".env")

	return Config{
		YouTubeAPIKey: mustEnv("YOUTUBE_API_KEY"),
		GeminiAPIKey:  mustEnv("GEMINI_API_KEY"),
		ChannelID:     mustEnv("YOUTUBE_CHANNEL_ID"),

		StateBucket: getEnv("STATE_BUCKET", ""),
		StateFile:   getEnv("STATE_FILE", "logs/last_video.txt"),

		EmailFrom:    getEnv("EMAIL_FROM", ""),
		EmailTo:      getEnv("EMAIL_TO", ""),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Required environment variable not set: %s", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
