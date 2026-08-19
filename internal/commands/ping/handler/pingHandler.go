package handler

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type PingHandler struct {
	Name        string
	Options     []*discordgo.ApplicationCommandOption
	Description string
}

func NewPingHandler() *PingHandler {
	return &PingHandler{
		Name:        "ping",
		Description: "Responds with Pong!",
		Options:     []*discordgo.ApplicationCommandOption{},
	}
}

func (p *PingHandler) GetName() string {
	return p.Name
}

func (p *PingHandler) GetDescription() string {
	return p.Description
}

func (p *PingHandler) GetOptions() []*discordgo.ApplicationCommandOption {
	return p.Options
}

func (p *PingHandler) Exec(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
	if err != nil {
		log.Println("Error responding to ping command:", err)
		return
	}

	username := "unknown"
	alias := "unknown"
	if i.Interaction.Member != nil && i.Interaction.Member.User != nil {
		alias = i.Interaction.Member.User.DisplayName()
		username = i.Interaction.Member.User.Username
	} else if i.Interaction.User != nil {
		alias = i.Interaction.User.Username
		username = i.Interaction.User.Username
	}

	log.Println(alias + " (" + username + ") executed the ping command successfully.")
}
