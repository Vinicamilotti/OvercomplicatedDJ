package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

func playStream(ctx context.Context, vc *discordgo.VoiceConnection, youtubeURL string) error {
	log.Printf("playStream: resolving %s", youtubeURL)

	resolveCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(resolveCtx, "yt-dlp",
		"-f", "bestaudio",
		"-j",
		"--no-playlist",
		youtubeURL,
	)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("yt-dlp resolve: %w", err)
	}

	var info struct {
		URL         string            `json:"url"`
		HTTPHeaders map[string]string `json:"http_headers"`
	}
	if err := json.Unmarshal(output, &info); err != nil {
		return fmt.Errorf("yt-dlp parse: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "yt-audio-*")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	log.Printf("playStream: downloading audio...")
	req, err := http.NewRequestWithContext(ctx, "GET", info.URL, nil)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("http request: %w", err)
	}
	for k, v := range info.HTTPHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("http download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("http status %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("download write: %w", err)
	}
	tmpFile.Close()

	st, _ := os.Stat(tmpPath)
	log.Printf("playStream: downloaded %d bytes, encoding...", st.Size())

	options := *dca.StdEncodeOptions
	options.RawOutput = true

	session, err := dca.EncodeFile(tmpPath, &options)
	if err != nil {
		return fmt.Errorf("dca EncodeFile: %w", err)
	}
	defer session.Cleanup()

	frames := 0
	for {
		frame, err := session.OpusFrame()
		if err != nil {
			if err == io.EOF {
				log.Printf("playStream: finished, sent %d frames", frames)
				return nil
			}
			return fmt.Errorf("OpusFrame: %w", err)
		}
		if frames == 0 {
			log.Printf("playStream: first frame len=%d, sending to OpusSend", len(frame))
		}
		select {
		case vc.OpusSend <- frame:
			frames++
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}