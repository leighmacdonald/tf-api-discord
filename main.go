package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/leighmacdonald/discordgo-lipstick/bot"
	"github.com/leighmacdonald/tf-api-discord/tfapi"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api, errAPI := tfapi.New(os.Getenv("TFAPI_URL"), &http.Client{Timeout: time.Second * 20})
	if errAPI != nil {
		return errAPI
	}

	bot, errBot := bot.New(bot.Opts{
		Token:     os.Getenv("DISCORD_TOKEN"),
		AppID:     os.Getenv("DISCORD_APP_ID"),
		GuildID:   os.Getenv("DISCORD_GUILD_ID"),
		UserAgent: "tf-api-discord (https://github.com/leighmacdonald/tf-api-discord)",
	})
	if errBot != nil {
		return errBot
	}
	defer bot.Close()

	if errRegister := registerCommands(ctx, bot, api); errRegister != nil {
		return errRegister
	}

	if errStart := bot.Start(ctx); errStart != nil {
		return errStart
	}

	<-ctx.Done()

	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("Exited on error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
