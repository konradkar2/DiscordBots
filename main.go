package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/konradkar2/marcus_discord/domain"
	"github.com/konradkar2/marcus_discord/infrastracture/mongo"
	// "go.mongodb.org/mongo-driver/v2/bson"
	// "go.mongodb.org/mongo-driver/v2/mongo"
	// "go.mongodb.org/mongo-driver/v2/mongo/options"
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

func (bot * Bot) startScheduler(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(d)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				fmt.Printf("tick!\n")
				bot.processDue(ctx)
			}
		}
	}()
}

type Bot struct{
	notesRepo domain.NotesRepository
	usersRepo domain.UsersRepository
	subsRepo domain.SubscriptionRepository
	dcSession * discordgo.Session
}

func main() {
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
	sub_repository   := infrastracture_mongo.NewSubscriptionRepository(mongo_driver)

	bot := Bot {
		notesRepo: notes_repository,
		usersRepo: users_repository,
		subsRepo: sub_repository,
		dcSession: dcSession,
	}

	dcSession.AddHandler(bot.messageCreate)

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
	bot.startScheduler(ctx, time.Second * 5)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dcSession.Close()
}

func (bot * Bot) processDue(ctx context.Context) {
	subs, err := bot.subsRepo.FindDue(ctx, time.Now())
	if err != nil {
		log.Println("find due error:", err)
		return
	}

	for _, sub := range subs {
		bot.sendNoteToUser(ctx, sub.UserId)

		next := sub.NextSendAt.Add(24 * time.Second)
		err := bot.subsRepo.UpdateSubscription(ctx, sub.UserId, next)
		if err != nil {
			log.Println("update error:", err)
		}
	}
}

func (bot * Bot) sendNoteToUser(ctx context.Context, userId string) error {
	note, err := bot.notesRepo.GetRandom(ctx)
	if err != nil {
		log.Printf("failed get note %v", err)
	}
	
	return SendDM(bot.dcSession, userId, note.Content)
}

func SendDM(s *discordgo.Session, userID string, msg string) error {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(channel.ID, msg)
	return err
}

func (bot * Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	//isPrivateMessage := (m.GuildID == "");

	ctx := context.TODO()

	bot.usersRepo.Insert(ctx,
		domain.User{
			UserId:   m.Author.ID,
			UserName: m.Author.Username})

	if m.Content == "subscribe" {
		nextSchedule := time.Now().Add(24 * time.Second)
		err := bot.subsRepo.UpdateSubscription(ctx, m.Author.ID, nextSchedule)

		if err != nil {
			log.Printf("failed UpdateSubscription %v", err)
		}
	}

	note, err := bot.notesRepo.GetRandom(ctx)
	if err != nil {
		log.Printf("failed get note %v", err)
	}

	s.ChannelMessageSend(m.ChannelID, note.Content)
}
