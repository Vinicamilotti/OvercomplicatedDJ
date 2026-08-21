package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vinicamilotti/OvercomplicatedDJ/configs"
	clearHandler "github.com/Vinicamilotti/OvercomplicatedDJ/internal/commands/clear/handler"
	pingHandler "github.com/Vinicamilotti/OvercomplicatedDJ/internal/commands/ping/handler"
	playHandler "github.com/Vinicamilotti/OvercomplicatedDJ/internal/commands/play/handler"
	queueCmdHandler "github.com/Vinicamilotti/OvercomplicatedDJ/internal/commands/queue/handler"
	skipHandler "github.com/Vinicamilotti/OvercomplicatedDJ/internal/commands/skip/handler"
	audioImpl "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/audio/impl"
	discRouter "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/discord/router"
	lava "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/lavalink"
	queueImpl "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/impl"
	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/yt"

	"github.com/bwmarrin/discordgo"
)

func main() {
	cfg, err := configs.LoadConfigs()
	if err != nil {
		panic(err)
	}

	discSession, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		panic(err)
	}

	router := discRouter.NewRouter(discSession)
	discSession.AddHandler(router.OnInteractionCreate)

	err = discSession.Open()
	if err != nil {
		panic(err)
	}
	defer discSession.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	queueSvc := queueImpl.NewQueueService()

	ll, err := lava.New(cfg, discSession.State.User.ID)
	if err != nil {
		panic(fmt.Errorf("lavalink: %w", err))
	}

	discSession.AddHandler(func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
		if e.UserID != discSession.State.User.ID {
			return
		}
		ll.OnVoiceStateUpdate(context.TODO(), e.GuildID, e.ChannelID, e.SessionID)
	})
	discSession.AddHandler(func(s *discordgo.Session, e *discordgo.VoiceServerUpdate) {
		ll.OnVoiceServerUpdate(context.TODO(), e.GuildID, e.Token, e.Endpoint)
	})

	playerSvc := audioImpl.NewPlayerService(discSession, ll, queueSvc)
	ytProvider := yt.NewYouTubeProvider(ll)

	err = router.AddCommand(pingHandler.NewPingHandler())
	if err != nil {
		panic(err)
	}
	err = router.AddCommand(playHandler.NewPlayHandler(queueSvc, playerSvc, ytProvider))
	if err != nil {
		panic(err)
	}
	err = router.AddCommand(skipHandler.NewSkipHandler(playerSvc))
	if err != nil {
		panic(err)
	}
	err = router.AddCommand(queueCmdHandler.NewQueueHandler(queueSvc))
	if err != nil {
		panic(err)
	}
	err = router.AddCommand(clearHandler.NewClearHandler(queueSvc, playerSvc))
	if err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	router.RemoveCommands()
}