package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/bwmarrin/discordgo"
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

	log.Printf("playStream: streaming audio...")
	req, err := http.NewRequestWithContext(ctx, "GET", info.URL, nil)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	for k, v := range info.HTTPHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %d", resp.StatusCode)
	}

	log.Printf("playStream: starting ffmpeg encode...")
	ffmpeg := exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0",
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "opus",
		"-flush_packets", "1",
		"-vbr", "on",
		"-compression_level", "10",
		"-ar", "48000",
		"-ac", "2",
		"-b:a", "64000",
		"-application", "audio",
		"-frame_duration", "20",
		"-packet_loss", "1",
		"pipe:1",
	)
	ffmpeg.Stdin = resp.Body

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}

	if err := ffmpeg.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	buf := make([]byte, 4096)
	frames := 0

	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			packet := make([]byte, n)
			copy(packet, buf[:n])

			if len(packet) < 20 {
				continue
			}

			if frames == 0 {
				log.Printf("playStream: first frame len=%d", len(packet))
			}

			select {
			case vc.OpusSend <- packet:
				frames++
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		if err != nil {
			if err == io.EOF {
				ffmpeg.Wait()
				log.Printf("playStream: finished, sent %d frames", frames)
				return nil
			}
			return fmt.Errorf("ffmpeg stdout: %w", err)
		}
	}
}