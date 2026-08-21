package configs

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken     string `env:"DISCORD_TOKEN"`
	LavalinkAddr     string `env:"LAVALINK_ADDR"`
	LavalinkPassword string `env:"LAVALINK_PASSWORD"`
}

var Configs *Config

func LoadConfigs() (*Config, error) {
	godotenv.Load() // Load environment variables from .env file
	cfg := &Config{}
	err := env.Parse(cfg)
	Configs = cfg
	return Configs, err
}
