package ports

type PlayerInterface interface {
	Play(guildID, voiceChannelID string) error
	Skip(guildID string) error
	Stop(guildID string) error
	IsPlaying(guildID string) bool
}