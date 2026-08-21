package yt

import (
	"context"
	"fmt"
	"strings"
	"time"

	lava "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/lavalink"
	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/media/entities"
)

type YouTubeProvider struct {
	ll *lava.Client
}

func NewYouTubeProvider(ll *lava.Client) *YouTubeProvider {
	return &YouTubeProvider{ll: ll}
}

func (p *YouTubeProvider) Supports(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

func (p *YouTubeProvider) Search(query string) (*entities.Track, error) {
	searchQuery := query
	if !strings.HasPrefix(query, "http") {
		searchQuery = "ytsearch:" + query
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tracks, err := p.ll.LoadTracks(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no results for: %s", query)
	}

	t := tracks[0]
	duration := int(t.Info.Length / 1000)

	uri := ""
	if t.Info.URI != nil {
		uri = *t.Info.URI
	}

	return &entities.Track{
		URL:      uri,
		Title:    t.Info.Title,
		Platform: "youtube",
		Duration: duration,
	}, nil
}