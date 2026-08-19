package ports

import (
	"github.com/bwmarrin/discordgo"
)

type Router interface {
	AddCommand(command DiscordCommandInterface) error
	OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate)
	RemoveCommands()
}
