package impl

import (
	"context"
	"log"
	"sync"

	audioPorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/audio/ports"
	lava "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/lavalink"
	queuePorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/ports"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v4/disgolink"
	lv "github.com/disgoorg/disgolink/v4/lavalink"
)

type guildPlayer struct {
	playing bool
}

type PlayerService struct {
	mu       sync.Mutex
	players  map[string]*guildPlayer
	session  *discordgo.Session
	ll       *lava.Client
	queueSvc queuePorts.QueueServiceInterface
}

func NewPlayerService(
	s *discordgo.Session,
	ll *lava.Client,
	queueSvc queuePorts.QueueServiceInterface,
) *PlayerService {
	ps := &PlayerService{
		players:  make(map[string]*guildPlayer),
		session:  s,
		ll:       ll,
		queueSvc: queueSvc,
	}

	ll.OnTrackEnd(func(guildID string, track lv.Track, reason lv.TrackEndReason) {
		log.Printf("Player: track ended in guild=%s reason=%s", guildID, reason)
		ps.mu.Lock()
		gp := ps.players[guildID]
		ps.mu.Unlock()

		if gp != nil && gp.playing {
			ps.playNext(guildID)
		}
	})

	return ps
}

func (ps *PlayerService) Play(guildID, voiceChannelID string) error {
	ps.mu.Lock()
	gp, exists := ps.players[guildID]
	if exists && gp.playing {
		ps.mu.Unlock()
		return nil
	}
	ps.players[guildID] = &guildPlayer{playing: true}
	ps.mu.Unlock()

	err := ps.session.ChannelVoiceJoinManual(guildID, voiceChannelID, false, false)
	if err != nil {
		log.Println("Failed to send voice state update:", err)
		ps.setPlaying(guildID, false)
		return err
	}

	ps.playNext(guildID)
	return nil
}

func (ps *PlayerService) playNext(guildID string) {
	queue := ps.queueSvc.GetOrCreate(guildID)
	entry, ok := queue.Next()
	if !ok {
		log.Println("Player: queue empty, stopping")
		ps.Stop(guildID)
		return
	}

	log.Printf("Player: loading track: %s", entry.Track.URL)
	tracks, err := ps.ll.LoadTracks(context.Background(), entry.Track.URL)
	if err != nil || len(tracks) == 0 {
		log.Printf("Player: failed to load track: %v", err)
		ps.playNext(guildID)
		return
	}

	player := ps.ll.Player(guildID)
	err = player.Update(context.Background(), disgolink.WithTrack(tracks[0]))
	if err != nil {
		log.Printf("Player: failed to play: %v", err)
		ps.playNext(guildID)
		return
	}

	log.Printf("Player: now playing: %s", entry.Track.Title)
}

func (ps *PlayerService) Skip(guildID string) error {
	log.Printf("Player: skipping in guild=%s", guildID)
	ps.playNext(guildID)
	return nil
}

func (ps *PlayerService) Stop(guildID string) error {
	log.Printf("Player: stopping guild=%s", guildID)
	ps.setPlaying(guildID, false)
	ps.ll.DestroyPlayer(guildID)
	ps.session.ChannelVoiceJoinManual(guildID, "", false, false)
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
	gp := ps.players[guildID]
	if gp == nil {
		if playing {
			ps.players[guildID] = &guildPlayer{playing: true}
		}
		return
	}
	gp.playing = playing
	if !playing {
		delete(ps.players, guildID)
	}
}

var _ audioPorts.PlayerInterface = (*PlayerService)(nil)