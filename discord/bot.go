package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

var (
	ErrConfig         = errors.New("configuration error")
	ErrCommandInvalid = errors.New("command invalid")
	ErrSession        = errors.New("failed to start session")
	ErrCommandSend    = errors.New("failed to send response")
)

type Handler func(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) (*discordgo.MessageEmbed, error)

type Command struct {
	Command *discordgo.ApplicationCommand
	Handler Handler
}

type Opts struct {
	AppID   string
	GuildID string
	Token   string
}

func New(opts Opts) (*Bot, error) {
	if opts.AppID == "" {
		return nil, fmt.Errorf("%w: invalid DISCORD_APP_ID", ErrConfig)
	}

	if opts.Token == "" {
		return nil, fmt.Errorf("%w: invalid DISCORD_TOKEN", ErrConfig)
	}

	bot := &Bot{appID: opts.AppID, guildID: opts.GuildID, commandHandlers: make(map[string]Handler)}
	session, errSession := discordgo.New("Bot " + opts.Token)
	if errSession != nil {
		return nil, errors.Join(errSession, ErrConfig)
	}
	session.UserAgent = "tf-api-discord (https://github.com/leighmacdonald/tf-api-discord)"
	session.Identify.Intents |= discordgo.IntentsGuildMessages
	session.Identify.Intents |= discordgo.IntentMessageContent
	session.Identify.Intents |= discordgo.IntentGuildMembers

	session.AddHandler(bot.onReady)

	bot.session = session
	bot.session.AddHandler(bot.onConnect)
	bot.session.AddHandler(bot.onDisconnect)
	bot.session.AddHandler(bot.onInteractionCreate)
	return bot, nil
}

type Bot struct {
	appID              string
	guildID            string
	session            *discordgo.Session
	commandHandlers    map[string]Handler
	commands           []*discordgo.ApplicationCommand
	running            atomic.Bool
	registeredCommands []*discordgo.ApplicationCommand
}

func (b *Bot) Start(ctx context.Context) error {
	if b.running.Load() {
		return nil
	}

	b.running.Store(true)

	if errStart := b.session.Open(); errStart != nil {
		return errors.Join(errStart, ErrSession)
	}

	return nil
}

func (b *Bot) Close() {
	if err := b.session.Close(); err != nil {
		slog.Error("failed to close discord session cleanly", slog.String("error", err.Error()))
	}
}

func (b *Bot) MustRegisterHandler(cmd string, handler Handler, appCommand *discordgo.ApplicationCommand) {
	_, found := b.commandHandlers[cmd]
	if found {
		panic(ErrCommandInvalid)
	}
	for _, cmd := range b.commands {
		if cmd.Name == appCommand.Name {
			panic(ErrCommandInvalid)
		}
	}

	b.commandHandlers[cmd] = handler
	b.commands = append(b.commands, appCommand)
}

func (b *Bot) onReady(session *discordgo.Session, _ *discordgo.Ready) {
	slog.Info("Logged in successfully", slog.String("name", session.State.User.Username), slog.String("discriminator", session.State.User.Discriminator))
}

func (b *Bot) onDisconnect(_ *discordgo.Session, _ *discordgo.Disconnect) {
	slog.Info("Discord state changed", slog.String("state", "disconnected"))
}

func (b *Bot) onInteractionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var (
		data    = interaction.ApplicationCommandData()
		command = data.Name
	)

	if handler, handlerFound := b.commandHandlers[command]; handlerFound {
		// sendPreResponse should be called for any commands that call external services or otherwise
		// could not return a response instantly. discord will time out commands that don't respond within a
		// very short timeout windows, ~2-3 seconds.
		initialResponse := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Calculating numberwang...",
			},
		}

		if errRespond := session.InteractionRespond(interaction.Interaction, initialResponse); errRespond != nil {
			if _, errFollow := session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
				Content: errRespond.Error(),
			}); errFollow != nil {
				slog.Error("Failed sending error response for interaction", slog.String("error", errFollow.Error()))
			}

			return
		}

		commandCtx, cancelCommand := context.WithTimeout(context.TODO(), time.Second*30)
		defer cancelCommand()

		response, errHandleCommand := handler(commandCtx, session, interaction)
		if errHandleCommand != nil || response == nil {
			if _, errFollow := session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{&discordgo.MessageEmbed{Title: "Error", Description: errHandleCommand.Error()}},
			}); errFollow != nil {
				slog.Error("Failed sending error response for interaction", slog.String("error", errFollow.Error()))
			}

			return
		}

		if sendSendResponse := b.sendInteractionResponse(session, interaction.Interaction, response); sendSendResponse != nil {
			slog.Error("Failed sending success response for interaction", slog.String("error", sendSendResponse.Error()))
		}
	}
}

func (b *Bot) sendInteractionResponse(session *discordgo.Session, interaction *discordgo.Interaction, response *discordgo.MessageEmbed) error {
	resp := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{response},
	}

	_, errResponseErr := session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
		Embeds: &resp.Embeds,
	})

	if errResponseErr != nil {
		if _, errResp := session.FollowupMessageCreate(interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong",
		}); errResp != nil {
			return errors.Join(errResp, ErrCommandSend)
		}

		return nil
	}

	return nil
}

func (b *Bot) onConnect(_ *discordgo.Session, _ *discordgo.Connect) {
	slog.Info("Discord state changed", slog.String("state", "connected"))

	if errRegister := b.overwriteCommands(); errRegister != nil {
		slog.Error("Failed to register discord slash commands", slog.String("error", errRegister.Error()))
	}
}

func (b *Bot) overwriteCommands() error {
	var slashCommands []*discordgo.ApplicationCommand
	for _, cmd := range b.commands {
		slashCommands = append(slashCommands, cmd)
	}

	// When guildID is empty, it registers the commands globally instead of per guild.
	commands, errBulk := b.session.ApplicationCommandBulkOverwrite(b.appID, b.guildID, slashCommands)
	if errBulk != nil {
		return errors.Join(errBulk, ErrCommandInvalid)
	}

	b.registeredCommands = commands

	return nil
}
