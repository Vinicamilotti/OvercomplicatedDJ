module github.com/Vinicamilotti/OvercomplicatedDJ

go 1.26.5

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/caarlos0/env/v11 v11.4.1
	github.com/joho/godotenv v1.5.1
	github.com/jonas747/dca v0.0.0-20210930103944-155f5e5f0cc7
)

require (
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jonas747/ogg v0.0.0-20161220051205-b4f6f4cf3757 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace github.com/bwmarrin/discordgo => github.com/Vinicamilotti/discordgo-fork v0.0.0-20260819175757-ad1693f34a15

replace github.com/jonas747/dca => github.com/Vinicamilotti/dca v0.0.0-20260819180326-d044648e2586
