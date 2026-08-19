package handler

import (
	"fmt"
	"log"
	"strings"

	mediaEntities "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/media/entities"
	mediaPorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/media/ports"
	audioPorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/audio/ports"
	queueEntities "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/entities"
	queuePorts "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/ports"

	"github.com/bwmarrin/discordgo"
)

type PlayHandler struct {
	Name         string
	Description  string
	Options      []*discordgo.ApplicationCommandOption
	QueueService queuePorts.QueueServiceInterface
	Player       audioPorts.PlayerInterface
	Provider     mediaPorts.MediaProviderInterface
}

func NewPlayHandler(
	queueSvc queuePorts.QueueServiceInterface,
	player audioPorts.PlayerInterface,
	provider mediaPorts.MediaProviderInterface,
) *PlayHandler {
	return &PlayHandler{
		Name:        "play",
		Description: "Adiciona uma musica a fila",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "musica",
				Description: "Nome ou URL da musica",
				Required:    true,
			},
		},
		QueueService: queueSvc,
		Player:       player,
		Provider:     provider,
	}
}

func (h *PlayHandler) GetName() string                                   { return h.Name }
func (h *PlayHandler) GetDescription() string                            { return h.Description }
func (h *PlayHandler) GetOptions() []*discordgo.ApplicationCommandOption { return h.Options }

func (h *PlayHandler) Exec(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := i.ApplicationCommandData().Options[0].StringValue()
	guildID := i.Interaction.GuildID
	log.Printf("Play command: query=%s guild=%s", query, guildID)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	var track *mediaEntities.Track
	var err error

	if strings.HasPrefix(query, "http") {
		if !h.Provider.Supports(query) {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: strPtr("URL não suportada. Por enquanto só YouTube."),
			})
			return
		}
		track, err = h.Provider.Search(query)
	} else {
		track, err = h.Provider.Search(query)
	}

	if err != nil {
		log.Printf("Play command: search failed: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("Erro ao buscar: %s", err.Error())),
		})
		return
	}

	requester := getUsername(i)
	entry := queueEntities.QueueEntry{
		Track:     *track,
		Requester: requester,
	}
	h.QueueService.GetOrCreate(guildID).Add(entry)

	if h.Player.IsPlaying(guildID) {
		pos := h.QueueService.GetOrCreate(guildID).Len()
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("Adicionado a fila (#%d): **%s**", pos, track.Title)),
		})
		return
	}

	voiceChannelID := getUserVoiceChannel(s, guildID, getMemberUserID(i))
	if voiceChannelID == "" {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Entre em um canal de voz primeiro!"),
		})
		return
	}

	go func() {
		if err := h.Player.Play(guildID, voiceChannelID); err != nil {
			s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Erro ao tocar: %s", err.Error()))
		}
	}()

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: strPtr(fmt.Sprintf("Tocando agora: **%s**", track.Title)),
	})
}

func strPtr(s string) *string {
	return &s
}

func getUserVoiceChannel(s *discordgo.Session, guildID, userID string) string {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return ""
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID
		}
	}
	return ""
}

func getUsername(i *discordgo.InteractionCreate) string {
	if i.Interaction.Member != nil && i.Interaction.Member.User != nil {
		return i.Interaction.Member.User.Username
	}
	if i.Interaction.User != nil {
		return i.Interaction.User.Username
	}
	return "unknown"
}

func getMemberUserID(i *discordgo.InteractionCreate) string {
	if i.Interaction.Member != nil && i.Interaction.Member.User != nil {
		return i.Interaction.Member.User.ID
	}
	if i.Interaction.User != nil {
		return i.Interaction.User.ID
	}
	return ""
}