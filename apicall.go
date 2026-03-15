package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

const geminiModel = "gemini-3-flash-preview"

type VideoInfo struct {
	ID          string
	Title       string
	PublishedAt string
	URL         string
}

type PromptConfig struct {
	SummaryPrompt string `yaml:"summary_prompt"`
}

func loadSummaryPrompt(video *VideoInfo) (string, error) {
	data, err := os.ReadFile("prompt.yaml")
	if err != nil {
		return "", fmt.Errorf("read prompt.yaml: %w", err)
	}

	var cfg PromptConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse prompt.yaml: %w", err)
	}
	if cfg.SummaryPrompt == "" {
		return "", fmt.Errorf("prompt.yaml missing summary_prompt")
	}

	return fmt.Sprintf(cfg.SummaryPrompt, video.Title, video.URL), nil
}

func getUploadsPlaylistID(apiKey, channelID string) (string, error) {
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/channels?part=contentDetails&id=%s&key=%s",
		channelID, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get channels: %w", err)
	}

	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ContentDetails struct {
				RelatedPlaylists struct {
					Uploads string `json:"uploads"`
				} `json:"relatedPlaylists"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode channels: %w", err)
	}
	if len(result.Items) == 0 {
		return "", fmt.Errorf("channel not found: %s", channelID)
	}

	return result.Items[0].ContentDetails.RelatedPlaylists.Uploads, nil
}

func getLatestVideo(apiKey, playlistID string) (*VideoInfo, error) {
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=1&playlistId=%s&key=%s",
		playlistID, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get playlist: %w", err)
	}

	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Snippet struct {
				Title       string `json:"title"`
				PublishedAt string `json:"publishedAt"`
				ResourceID  struct {
					VideoID string `json:"videoId"`
				} `json:"resourceId"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode playlist: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no videos found in playlist")
	}

	item := result.Items[0]
	id := item.Snippet.ResourceID.VideoID

	return &VideoInfo{
		ID:          id,
		Title:       item.Snippet.Title,
		PublishedAt: item.Snippet.PublishedAt,
		URL:         "https://www.youtube.com/watch?v=" + id,
	}, nil
}

func summarizeVideo(apiKey string, video *VideoInfo) (string, error) {
	prompt, err := loadSummaryPrompt(video)
	if err != nil {
		return "", err
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					// Pass the YouTube URL as a native video input — Gemini reads it directly
					{"fileData": map[string]string{
						"mimeType": "video/mp4",
						"fileUri":  video.URL,
					}},
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.3,
		},
	}

	body, _ := json.Marshal(reqBody)
	apiURL := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		geminiModel, apiKey,
	)

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini http post: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("gemini parse: %w\nbody: %s", err, string(raw))
	}
	if result.Error != nil {
		return "", fmt.Errorf("gemini api error: %s", result.Error.Message)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty response, body: %s", string(raw))
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
