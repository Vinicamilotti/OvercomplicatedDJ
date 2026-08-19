package router

import (
	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/discord/domain/ports"
	"github.com/bwmarrin/discordgo"
)

type Router struct {
	session           *discordgo.Session
	Commands          map[string]ports.DiscordCommandInterface
	registeredCommands []*discordgo.ApplicationCommand
}

func NewRouter(s *discordgo.Session) *Router {
	return &Router{
		session:  s,
		Commands: make(map[string]ports.DiscordCommandInterface),
	}
}

func (r *Router) getGuildID() string {
	return "1515795104474988695"
}

func (r *Router) AddCommand(command ports.DiscordCommandInterface) error {
	r.Commands[command.GetName()] = command
	cmd, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, r.getGuildID(), &discordgo.ApplicationCommand{
		Name:        command.GetName(),
		Description: command.GetDescription(),
		Options:     command.GetOptions(),
	})
	if err != nil {
		return err
	}
	r.registeredCommands = append(r.registeredCommands, cmd)
	return nil
}

func (r *Router) RemoveCommands() {
	for _, cmd := range r.registeredCommands {
		r.session.ApplicationCommandDelete(r.session.State.User.ID, r.getGuildID(), cmd.ID)
	}
}

func (r *Router) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if command, ok := r.Commands[i.ApplicationCommandData().Name]; ok {
		command.Exec(s, i)
	}
}
