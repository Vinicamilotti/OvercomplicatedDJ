package impl

import (
	"context"
	"log"
	"sync"
	"time"

	audioPorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/audio/ports"
	queuePorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/ports"

	"github.com/bwmarrin/discordgo"
)

type guildPlayer struct {
	cancel  context.CancelFunc
	playing bool
}

type PlayerService struct {
	mu       sync.Mutex
	players  map[string]*guildPlayer
	session  *discordgo.Session
	queueSvc queuePorts.QueueServiceInterface
}

func NewPlayerService(
	s *discordgo.Session,
	queueSvc queuePorts.QueueServiceInterface,
) *PlayerService {
	return &PlayerService{
		players:  make(map[string]*guildPlayer),
		session:  s,
		queueSvc: queueSvc,
	}
}

func (ps *PlayerService) Play(guildID, voiceChannelID string) error {
	ps.mu.Lock()
	gp, exists := ps.players[guildID]
	if exists && gp.playing {
		ps.mu.Unlock()
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	ps.players[guildID] = &guildPlayer{cancel: cancel, playing: true}
	ps.mu.Unlock()

	go ps.playLoop(ctx, guildID, voiceChannelID)
	return nil
}

func (ps *PlayerService) playLoop(ctx context.Context, guildID, voiceChannelID string) {
	log.Printf("playLoop: joining voice channel %s in guild %s", voiceChannelID, guildID)
	vc, err := ps.session.ChannelVoiceJoin(context.Background(), guildID, voiceChannelID, false, false)
	if err != nil {
		log.Println("Failed to join voice channel:", err)
		ps.setPlaying(guildID, false)
		return
	}
	defer vc.Disconnect(context.Background())
	log.Println("playLoop: connected to voice channel")

	queue := ps.queueSvc.GetOrCreate(guildID)

	for {
		entry, ok := queue.Next()
		if !ok {
			log.Println("playLoop: queue empty, waiting 30s...")
			select {
			case <-time.After(30 * time.Second):
				if queue.Len() == 0 {
					ps.setPlaying(guildID, false)
					return
				}
				continue
			case <-ctx.Done():
				ps.setPlaying(guildID, false)
				return
			}
		}

		log.Printf("Playing: %s\n", entry.Track.Title)
		if err := playTestTone(ctx, vc); err != nil {
			if err == context.Canceled {
				return
			}
			log.Println("Stream error:", err)
		}
	}
}

func (ps *PlayerService) Skip(guildID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	gp, exists := ps.players[guildID]
	if !exists || !gp.playing {
		return nil
	}
	gp.cancel()
	return nil
}

func (ps *PlayerService) Stop(guildID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	gp, exists := ps.players[guildID]
	if !exists {
		return nil
	}
	if gp.playing {
		gp.cancel()
	}
	gp.playing = false
	return nil
}

func (ps *PlayerService) IsPlaying(guildID string) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	gp, exists := ps.players[guildID]
	return exists && gp.playing
}

func (ps *PlayerService) setPlaying(guildID string, playing bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if gp, exists := ps.players[guildID]; exists {
		gp.playing = playing
	}
}

var _ audioPorts.PlayerInterface = (*PlayerService)(nil)