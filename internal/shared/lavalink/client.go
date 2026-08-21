package lavalink

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Vinicamilotti/OvercomplicatedDJ/configs"
	"github.com/disgoorg/disgolink/v4/disgolink"
	ll "github.com/disgoorg/disgolink/v4/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type Client struct {
	client          *disgolink.Client
	node            *disgolink.Node
	userID          snowflake.ID
	trackEndHandler func(guildID string, track ll.Track, reason ll.TrackEndReason)
	mu              sync.Mutex
}

func New(cfg *configs.Config, sessUserID string) (*Client, error) {
	userID, err := strconv.ParseUint(sessUserID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	snowflakeID := snowflake.ID(userID)

	c := &Client{
		userID: snowflakeID,
	}

	c.client = disgolink.New(snowflakeID,
		disgolink.WithListenerFunc(func(e *disgolink.ReadyEvent) {
			log.Println("Lavalink: ready, session", e.SessionID)
		}),
		disgolink.WithListenerFunc(func(e *disgolink.PlayerTrackEndEvent) {
			c.mu.Lock()
			handler := c.trackEndHandler
			c.mu.Unlock()
			if handler != nil {
				handler(
					strconv.FormatUint(uint64(e.GuildID), 10),
					e.Track,
					e.Reason,
				)
			}
		}),
	)

	node, err := c.client.AddNode(context.Background(), disgolink.NodeConfig{
		Name:     "main",
		Address:  cfg.LavalinkAddr,
		Password: cfg.LavalinkPassword,
		Secure:   false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Lavalink: %w", err)
	}
	c.node = node
	log.Println("Lavalink: connected to node")

	return c, nil
}

func (c *Client) LoadTracks(ctx context.Context, query string) ([]ll.Track, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var result []ll.Track
	var loadErr error
	done := make(chan struct{})

	c.node.Rest.LoadTracksHandler(ctx, query, disgolink.NewTrackLoadingResultHandler(
		func(track ll.Track) {
			result = []ll.Track{track}
			close(done)
		},
		func(playlist ll.Playlist) {
			result = playlist.Tracks
			close(done)
		},
		func(tracks []ll.Track) {
			result = tracks
			close(done)
		},
		func() {
			loadErr = fmt.Errorf("no matches for: %s", query)
			close(done)
		},
		func(err error) {
			loadErr = fmt.Errorf("load failed: %w", err)
			close(done)
		},
	))

	select {
	case <-done:
		return result, loadErr
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) Player(guildID string) *disgolink.Player {
	id, _ := strconv.ParseUint(guildID, 10, 64)
	return c.client.Player(snowflake.ID(id))
}

func (c *Client) DestroyPlayer(guildID string) {
	id, _ := strconv.ParseUint(guildID, 10, 64)
	c.client.RemovePlayer(snowflake.ID(id))
}

func (c *Client) OnVoiceStateUpdate(ctx context.Context, guildID string, channelID string, sessionID string) {
	gid, _ := strconv.ParseUint(guildID, 10, 64)
	cid, err := strconv.ParseUint(channelID, 10, 64)
	var channelIDPtr *snowflake.ID
	if err == nil {
		id := snowflake.ID(cid)
		channelIDPtr = &id
	}
	c.client.OnVoiceStateUpdate(ctx, snowflake.ID(gid), channelIDPtr, sessionID)
}

func (c *Client) OnVoiceServerUpdate(ctx context.Context, guildID string, token string, endpoint string) {
	gid, _ := strconv.ParseUint(guildID, 10, 64)
	c.client.OnVoiceServerUpdate(ctx, snowflake.ID(gid), token, endpoint)
}

func (c *Client) OnTrackEnd(handler func(guildID string, track ll.Track, reason ll.TrackEndReason)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.trackEndHandler = handler
}