# Vidscribe

Vidscribe is a Go-based automation tool designed to monitor a specific YouTube channel for new uploads, automatically summarize the video using Google's Gemini LLM, and send a notification via email. It is designed to be easily deployed and scheduled as a Google Cloud Run Job.

## Features

- **YouTube Monitoring:** Fetches the latest video from a configured YouTube channel using the YouTube Data API.
- **State Management:** Keeps track of the last processed video ID to avoid duplicate notifications. Supports Google Cloud Storage for production or local files for testing.
- **AI-Powered Summarization:** Uses the Gemini API to generate concise summaries of new videos.
- **Email Notifications:** Automatically sends an email containing the video link and its AI-generated summary.

## Prerequisites

To run or deploy Vidscribe, you will need:
- Go 1.24+ installed (for local runs)
- Docker (for containerization)
- A YouTube Data API Key
- A Google Gemini API Key
- A Google Cloud Platform (GCP) Account (for deployment)
- SMTP credentials to send the emails

## Configuration

Vidscribe relies on environment variables for configuration. You can use a local `.env` file when developing locally.

### Environment Variables

| Variable | Description |
|---|---|
| `YOUTUBE_API_KEY` | YouTube Data API key. |
| `GEMINI_API_KEY` | Google Gemini API key used for summaries. |
| `YOUTUBE_CHANNEL_ID` | The YouTube Channel ID you want to monitor. |
| `STATE_FILE` | Folder storage for logs on last run (e.g. logs/last_video.txt) |
| `STATE_BUCKET` | GCP Storage Bucket to hold the state (used in prod). |
| `EMAIL_FROM` | The sender email address. |
| `EMAIL_TO` | The notification recipient email address. |
| `SMTP_HOST` | Hostname for outgoing mail (e.g. smtp.gmail.com). |
| `SMTP_PORT` | Port for the SMTP server (e.g. 587). |
| `SMTP_USER` | SMTP authentication username. |
| `SMTP_PASSWORD` | SMTP authentication password. |

## Running Locally

1. Create a `.env` file at the root of the project with the variables detailed above.

2. Run the application:
   ```bash
   go run .
   ```

## Deployment via Google Cloud

Vidscribe is built to run as a scheduled **Google Cloud Run Job**. 

The scheduling for this job was configured manually via the Google Cloud Console using Google Cloud Scheduler. It is set up with a **CRON expression to execute every other week on Friday** (bi-weekly).

## Logs & Output

When testing locally (or depending on your environment):
- Summaries in plain markdown are saved to `logs/summary_[DATE].md`.
- The latest video ID processed is maintained in `logs/last_video.txt` (when using local state management).
