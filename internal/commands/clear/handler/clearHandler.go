package handler

import (
	audioPorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/audio/ports"
	queuePorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/ports"

	"github.com/bwmarrin/discordgo"
)

type ClearHandler struct {
	Name         string
	Description  string
	Options      []*discordgo.ApplicationCommandOption
	QueueService queuePorts.QueueServiceInterface
	Player       audioPorts.PlayerInterface
}

func NewClearHandler(queueSvc queuePorts.QueueServiceInterface, player audioPorts.PlayerInterface) *ClearHandler {
	return &ClearHandler{
		Name:         "clear",
		Description:  "Limpa a fila de musicas",
		Options:      []*discordgo.ApplicationCommandOption{},
		QueueService: queueSvc,
		Player:       player,
	}
}

func (h *ClearHandler) GetName() string                                   { return h.Name }
func (h *ClearHandler) GetDescription() string                            { return h.Description }
func (h *ClearHandler) GetOptions() []*discordgo.ApplicationCommandOption { return h.Options }

func (h *ClearHandler) Exec(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.Interaction.GuildID

	h.Player.Stop(guildID)
	h.QueueService.GetOrCreate(guildID).Clear()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Fila limpa e playback parado.",
		},
	})
}