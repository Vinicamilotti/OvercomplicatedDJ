package ports

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordCommandInterface interface {
	GetName() string
	GetDescription() string
	GetOptions() []*discordgo.ApplicationCommandOption
	Exec(s *discordgo.Session, i *discordgo.InteractionCreate)
}
