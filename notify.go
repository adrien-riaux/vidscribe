package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"time"
)

func saveToFile(video *VideoInfo, summary string) error {
	filename := fmt.Sprintf("logs/summary_%s.md", time.Now().Format("2006-01-02"))
	content := fmt.Sprintf(
		"# %s\n\n**URL:** %s  \n**Published:** %s  \n**Summarized:** %s\n\n---\n\n%s\n",
		video.Title, video.URL, video.PublishedAt, time.Now().Format(time.RFC3339), summary,
	)

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}
	log.Printf("Saved: %s", filename)

	return nil
}

func sendEmail(cfg Config, video *VideoInfo, summary string) error {
	if cfg.EmailTo == "" || cfg.EmailFrom == "" || cfg.SMTPUser == "" || cfg.SMTPPassword == "" || cfg.SMTPHost == "" {
		log.Println("Email not fully configured (EMAIL_TO, EMAIL_FROM, SMTP_USER, SMTP_PASSWORD, or SMTP_HOST missing) — skipping.")
		return nil
	}

	msg := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: New Video Summary: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n"+
			"New video detected on your tracked channel!\n\n"+
			"Title: %s\nLink: %s\nPublished: %s\n\n---\n\n%s",
		cfg.EmailTo, cfg.EmailFrom, video.Title,
		video.Title, video.URL, video.PublishedAt, summary,
	)

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	if err := smtp.SendMail(
		cfg.SMTPHost+":"+cfg.SMTPPort,
		auth,
		cfg.EmailFrom,
		[]string{cfg.EmailTo},
		[]byte(msg),
	); err != nil {
		return fmt.Errorf("smtp: %w", err)
	}
	log.Printf("Email sent to %s", cfg.EmailTo)

	return nil
}
