module github.com/konradkar2/marcus_discord

go 1.25.0

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/konradkar2/marcus_discord/discord v0.0.0-00010101000000-000000000000
	github.com/konradkar2/marcus_discord/infrastracture/mongo v0.0.0-00010101000000-000000000000
	github.com/konradkar2/marcus_discord/render v0.0.0-00010101000000-000000000000
)

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/konradkar2/marcus_discord/domain v0.0.0-00010101000000-000000000000 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver/v2 v2.5.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/image v0.38.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/konradkar2/marcus_discord/infrastracture/mongo => ./infrastracture/mongo

replace github.com/konradkar2/marcus_discord/domain => ./domain

replace github.com/konradkar2/marcus_discord/discord => ./discord

replace github.com/konradkar2/marcus_discord/render => ./render
