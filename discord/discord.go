package discord

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/konradkar2/marcus_discord/domain"
)

type CommandID int

const (
	CommandUnknown CommandID = iota
	CommandSubscribe
	CommandUnsubscribe
)

type Command struct {
	Id      CommandID
	Cmd     string
	Help    string
	Value   string
	Default string
}

func (c *Command) String() string {
	str := c.Cmd + ": " + c.Help + "\n"
	return str
}

func parseCommand(cmd string) []Command {
	var parts = strings.Split(cmd, " ")

	arr := []Command{}

	println(arr)

	isCommand := true
	for _, str := range parts {
		println("part: " + str)
		for _, command := range Commands {
			if isCommand {
				if command.Cmd == str {
					arr = append(arr, command)
					isCommand = false
					break;
				}
			} else {
				arr[len(arr)-1].Value = str
				isCommand = true
				break;
			}

		}
	}
	return arr
}

var Commands = []Command{
	SubscribeCommand,
	UnsubscribeCommand,
}

var SubscribeCommand = Command{
	Id:  CommandSubscribe,
	Cmd: "--subscribe", Help: "subskrybuj na codzienne notatki <interwał w sekundach>",
}

var UnsubscribeCommand = Command{
	Id:  CommandUnsubscribe,
	Cmd: "--unsubscribe", Help: "anuluj subskrypcje na codzienne notatki",
}

type Bot struct {
	notesRepo domain.NotesRepository
	usersRepo domain.UsersRepository
	subsRepo  domain.SubscriptionRepository
	dcSession *discordgo.Session
}

func (bot *Bot) HandleCommands(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, commands []Command) {
	for _, command := range commands {
		if command.Id == CommandSubscribe {
			fmt.Printf("command value:%v", command.Value)
			seconds, err := strconv.Atoi(command.Value)
			if err != nil || seconds <= 0 {
				s.ChannelMessageSend(m.ChannelID, "Nieprawidłowy interwał. Użyj liczby sekund.")
			}

			var interval = time.Duration(seconds) * time.Second
			nextSchedule := time.Now().Add(interval)
			err = bot.subsRepo.UpdateSubscription(ctx, m.Author.ID, interval,nextSchedule)

			if err != nil {
				log.Printf("failed UpdateSubscription %v", err)
			}
		}
	}
}

func NewBot(notesRepo domain.NotesRepository,
	usersRepo domain.UsersRepository,
	subsRepo domain.SubscriptionRepository,
	dcSession *discordgo.Session) *Bot {
	return &Bot{
		notesRepo: notesRepo,
		usersRepo: usersRepo,
		subsRepo:  subsRepo,
		dcSession: dcSession,
	}
}

func (bot *Bot) StartScheduler(ctx context.Context, d time.Duration) {
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

func (bot *Bot) processDue(ctx context.Context) {
	subs, err := bot.subsRepo.FindDue(ctx, time.Now())
	if err != nil {
		log.Println("find due error:", err)
		return
	}

	for _, sub := range subs {
		bot.sendNoteToUser(ctx, sub.UserId)

		next := sub.NextSendAt.Add(sub.Interval)
		err := bot.subsRepo.UpdateSubscription(ctx, sub.UserId,sub.Interval, next)
		if err != nil {
			log.Println("update error:", err)
		}
	}
}

func (bot *Bot) sendNoteToUser(ctx context.Context, userId string) error {
	note, err := bot.notesRepo.GetRandom(ctx)
	if err != nil {
		log.Printf("failed get note %v", err)
	}

	return sendDM(bot.dcSession, userId, note.Content)
}

func sendDM(s *discordgo.Session, userID string, msg string) error {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(channel.ID, msg)
	return err
}

func (bot *Bot) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	println("handleMessage");

	//isPrivateMessage := (m.GuildID == "");

	ctx := context.TODO()

	bot.usersRepo.Insert(ctx,
		domain.User{
			UserId:   m.Author.ID,
			UserName: m.Author.Username})

	s.ChannelMessageSend(m.ChannelID, SubscribeCommand.String())

	var parsedCommands = parseCommand(m.Content)
	bot.HandleCommands(ctx, s, m, parsedCommands)

	note, err := bot.notesRepo.GetRandom(ctx)
	if err != nil {
		log.Printf("failed get note %v", err)
	}

	s.ChannelMessageSend(m.ChannelID, note.Content)
}
