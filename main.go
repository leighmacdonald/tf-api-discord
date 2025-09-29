package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leighmacdonalf/tf-api-discord/discord"
	"github.com/leighmacdonalf/tf-api-discord/tfapi"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api, errAPI := tfapi.New(os.Getenv("TFAPI_URL"), &http.Client{Timeout: time.Second * 20})
	if errAPI != nil {
		return errAPI
	}

	bot, errBot := discord.New(discord.Opts{
		Token:   os.Getenv("DISCORD_TOKEN"),
		AppID:   os.Getenv("DISCORD_APP_ID"),
		GuildID: os.Getenv("DISCORD_GUILD_ID"),
	})
	if errBot != nil {
		return errBot
	}
	defer bot.Close()

	registerCommands(bot, api)

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
