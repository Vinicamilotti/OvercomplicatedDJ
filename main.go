package main

import (
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
	discSession.LogLevel = discordgo.LogInformational

	router := discRouter.NewRouter(discSession)

	discSession.AddHandler(router.OnInteractionCreate)

	err = discSession.Open()
	if err != nil {
		panic(err)
	}
	defer discSession.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	queueSvc := queueImpl.NewQueueService()
	ytProvider := yt.NewYouTubeProvider()
	playerSvc := audioImpl.NewPlayerService(discSession, queueSvc)

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