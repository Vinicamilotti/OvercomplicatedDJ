package handler

import (
	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/audio/ports"
	"github.com/bwmarrin/discordgo"
)

type SkipHandler struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
	Player      ports.PlayerInterface
}

func NewSkipHandler(player ports.PlayerInterface) *SkipHandler {
	return &SkipHandler{
		Name:        "skip",
		Description: "Pula a musica atual",
		Options:     []*discordgo.ApplicationCommandOption{},
		Player:      player,
	}
}

func (h *SkipHandler) GetName() string                                   { return h.Name }
func (h *SkipHandler) GetDescription() string                            { return h.Description }
func (h *SkipHandler) GetOptions() []*discordgo.ApplicationCommandOption { return h.Options }

func (h *SkipHandler) Exec(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.Interaction.GuildID

	if !h.Player.IsPlaying(guildID) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Nao ha nada tocando no momento.",
			},
		})
		return
	}

	h.Player.Skip(guildID)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Musica pulada!",
		},
	})
}