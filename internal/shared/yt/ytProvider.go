package yt

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/media/entities"
)

type YouTubeProvider struct{}

func NewYouTubeProvider() *YouTubeProvider {
	return &YouTubeProvider{}
}

func (p *YouTubeProvider) Supports(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

func (p *YouTubeProvider) Search(query string) (*entities.Track, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp", fmt.Sprintf("ytsearch1:%s", query), "-j", "--no-playlist")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search failed: %w", err)
	}

	var result struct {
		ID         string  `json:"id"`
		Title      string  `json:"title"`
		WebpageURL string  `json:"webpage_url"`
		Duration   float64 `json:"duration"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	return &entities.Track{
		URL:      result.WebpageURL,
		Title:    result.Title,
		Platform: "youtube",
		Duration: int(result.Duration),
	}, nil
}

func (p *YouTubeProvider) GetStreamURL(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp", "-f", "bestaudio", "-g", url)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("yt-dlp get stream URL failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}