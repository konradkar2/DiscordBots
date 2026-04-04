package main

import (
	"context"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/konradkar2/marcus_discord/discord"
	infrastracture_mongo "github.com/konradkar2/marcus_discord/infrastracture/mongo"
	"github.com/konradkar2/marcus_discord/render"
)

var (
	Token      string
	MongodbURI string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&MongodbURI, "m", "", "MongoDB connection URI")
	flag.Parse()
}

func main() {
	renderDemo := true

	if renderDemo {
		opts := render.DefaultRenderOptions()
		opts.BackgroundPath = "background.jpg"
		opts.FontPath = "PlayfairDisplay-VariableFont_wght.ttf"
		opts.OutputPath = "quote.png"
		opts.TextColor = color.RGBA{250, 248, 242, 255}
		opts.AuthorColor = color.RGBA{220, 218, 210, 255}

		err := render.RenderQuoteCard(
			`Kochaj sztukę, której się nauczyłeś, i w niej znajdź spokój. Resztę zaś życia spędź jako człowiek, który z całej duszy zdał swe sprawy w ręce bogów, a z siebie nie czyni ani tyrana, ani sługi żadnego człowieka.`,
			"Marcus Aurelius",
			opts,
		)
		if err != nil {
			log.Fatal(err)
		}
	} else {

		dcSession, err := discordgo.New("Bot " + Token)
		if err != nil {
			fmt.Println("error creating Discord session,", err)
			return
		}

		mongo_driver, err := infrastracture_mongo.NewMongoDriver(MongodbURI, "marcusDatabase")
		if err != nil {
			log.Fatalf("failed to create MongoDriver: %v", err)
		}

		notes_repository := infrastracture_mongo.NewNotesRepository(mongo_driver)
		users_repository := infrastracture_mongo.NewUsersRepository(mongo_driver)
		sub_repository := infrastracture_mongo.NewSubscriptionRepository(mongo_driver)

		bot := discord.NewBot(notes_repository, users_repository, sub_repository, dcSession)

		dcSession.AddHandler(bot.HandleMessage)

		dcSession.Identify.Intents =
			discordgo.IntentsGuildMessages |
				discordgo.IntentsDirectMessages |
				discordgo.IntentsMessageContent

		err = dcSession.Open()
		if err != nil {
			fmt.Println("error opening connection,", err)
			return
		}

		ctx := context.TODO()
		bot.StartScheduler(ctx, time.Second*5)

		fmt.Println("Bot is now running.  Press CTRL-C to exit.")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc

		println("goodbye!")
		dcSession.Close()
	}
}
