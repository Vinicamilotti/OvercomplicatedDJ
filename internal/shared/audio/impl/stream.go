package impl

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/bwmarrin/discordgo"
)

func playStream(ctx context.Context, vc *discordgo.VoiceConnection, youtubeURL string) error {
	log.Printf("playStream: downloading and converting to mp3...")

	ytdlp := exec.CommandContext(ctx, "yt-dlp",
		"-x", "--audio-format", "mp3",
		"-o", "-",
		"--no-playlist",
		youtubeURL,
	)

	mp3out, err := ytdlp.StdoutPipe()
	if err != nil {
		return fmt.Errorf("yt-dlp stdout pipe: %w", err)
	}

	if err := ytdlp.Start(); err != nil {
		return fmt.Errorf("yt-dlp start: %w", err)
	}

	log.Printf("playStream: encoding mp3 to opus...")
	ffmpeg := exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0",
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "ogg",
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
	ffmpeg.Stdin = mp3out

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}

	if err := ffmpeg.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	reader := newOggReader(stdout)
	frames := 0

	for {
		packet, err := reader.readPacket()
		if err != nil {
			if err == io.EOF {
				ffmpeg.Wait()
				log.Printf("playStream: finished, sent %d frames", frames)
				return nil
			}
			return fmt.Errorf("ogg read: %w", err)
		}

		if len(packet) < 20 || packet[0]&0x80 == 0 {
			continue
		}

		if frames == 0 {
			log.Printf("playStream: first audio frame len=%d hex=%s",
				len(packet), hex.EncodeToString(packet[:min(16, len(packet))]))
		}

		select {
		case vc.OpusSend <- packet:
			frames++
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type oggReader struct {
	r         io.Reader
	buf       []byte
	pos       int
	end       int
	segments  []byte
	segIdx    int
	packetBuf []byte
}

func newOggReader(r io.Reader) *oggReader {
	return &oggReader{r: r, buf: make([]byte, 65536)}
}

func (o *oggReader) readPacket() ([]byte, error) {
	for {
		for o.segIdx < len(o.segments) {
			segLen := int(o.segments[o.segIdx])
			o.segIdx++

			if o.pos+segLen > o.end {
				if err := o.fill(); err != nil {
					return nil, err
				}
				if o.pos+segLen > o.end {
					return nil, io.ErrUnexpectedEOF
				}
			}

			o.packetBuf = append(o.packetBuf, o.buf[o.pos:o.pos+segLen]...)
			o.pos += segLen

			if segLen < 255 {
				packet := o.packetBuf
				o.packetBuf = nil
				return packet, nil
			}
		}

		if err := o.readPage(); err != nil {
			return nil, err
		}
	}
}

func (o *oggReader) readPage() error {
	if err := o.fill(); err != nil {
		return err
	}

	for o.end-o.pos >= 4 && string(o.buf[o.pos:o.pos+4]) != "OggS" {
		o.pos++
		if o.pos >= o.end {
			if err := o.fill(); err != nil {
				return err
			}
		}
	}

	if o.end-o.pos < 27 {
		return io.ErrUnexpectedEOF
	}

	hdr := o.buf[o.pos : o.pos+27]
	segCount := int(hdr[26])
	o.pos += 27

	for o.pos+segCount > o.end {
		if err := o.fill(); err != nil {
			return err
		}
	}

	o.segments = o.buf[o.pos : o.pos+segCount]
	o.segIdx = 0
	o.pos += segCount

	dataLen := 0
	for _, s := range o.segments {
		dataLen += int(s)
	}

	for o.pos+dataLen > o.end {
		if err := o.fill(); err != nil {
			return err
		}
		if dataLen > len(o.buf)-o.pos {
			return fmt.Errorf("ogg page too large: %d bytes", dataLen)
		}
	}

	return nil
}

func (o *oggReader) fill() error {
	if o.pos > 0 && o.pos < o.end {
		copy(o.buf, o.buf[o.pos:o.end])
		o.end -= o.pos
		o.pos = 0
	}

	if o.end >= len(o.buf) {
		return fmt.Errorf("ogg buffer full")
	}

	n, err := o.r.Read(o.buf[o.end:])
	if n > 0 {
		o.end += n
	}
	if err != nil {
		if err == io.EOF && o.end > o.pos {
			return nil
		}
		return err
	}
	return nil
}