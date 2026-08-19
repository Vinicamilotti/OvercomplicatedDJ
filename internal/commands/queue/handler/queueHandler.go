package handler

import (
	"fmt"
	"strings"

	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/ports"
	"github.com/bwmarrin/discordgo"
)

type QueueHandler struct {
	Name         string
	Description  string
	Options      []*discordgo.ApplicationCommandOption
	QueueService ports.QueueServiceInterface
}

func NewQueueHandler(queueSvc ports.QueueServiceInterface) *QueueHandler {
	return &QueueHandler{
		Name:         "queue",
		Description:  "Mostra a fila de musicas",
		Options:      []*discordgo.ApplicationCommandOption{},
		QueueService: queueSvc,
	}
}

func (h *QueueHandler) GetName() string                                   { return h.Name }
func (h *QueueHandler) GetDescription() string                            { return h.Description }
func (h *QueueHandler) GetOptions() []*discordgo.ApplicationCommandOption { return h.Options }

func (h *QueueHandler) Exec(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.Interaction.GuildID
	entries := h.QueueService.GetOrCreate(guildID).List()

	if len(entries) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "A fila esta vazia.",
			},
		})
		return
	}

	var sb strings.Builder
	sb.WriteString("**Fila de musicas:**\n")
	for idx, entry := range entries {
		sb.WriteString(fmt.Sprintf("%d. **%s** - %s (pedido por %s)\n",
			idx+1, entry.Track.Title, entry.Track.Platform, entry.Requester))
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: sb.String(),
		},
	})
}