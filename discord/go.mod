module github.com/konradkar2/marcus_discord/discord

go 1.25.0

replace github.com/konradkar2/marcus_discord/infrastracture/mongo => ../infrastracture/mongo

replace github.com/konradkar2/marcus_discord/domain => ../domain

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/konradkar2/marcus_discord/domain v0.0.0-00010101000000-000000000000
)

require (
	github.com/alecthomas/kong v1.14.0
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)
