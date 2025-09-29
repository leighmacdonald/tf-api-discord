package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/leighmacdonalf/tf-api-discord/discord"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot, errBot := discord.New(discord.Opts{
		Token:   os.Getenv("DISCORD_TOKEN"),
		AppID:   os.Getenv("DISCORD_APP_ID"),
		GuildID: os.Getenv("DISCORD_GUILD_ID"),
	})
	if errBot != nil {
		return errBot
	}
	defer bot.Close()

	if errStart := bot.Start(ctx); errStart != nil {
		return errStart
	}

	slog.Info("Running...")
	<-ctx.Done()

	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("exited on error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
