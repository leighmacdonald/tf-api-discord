package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/leighmacdonald/steamid/v4/steamid"
	"github.com/leighmacdonalf/tf-api-discord/discord"
)

func registerCommands(bot *discord.Bot) {
	defaultCtx := &[]discordgo.InteractionContextType{discordgo.InteractionContextBotDM}
	//modPerms := int(discordgo.PermissionBanMembers)
	userPerms := int64(discordgo.PermissionViewChannel)

	bot.MustRegisterHandler("check", onCheck, &discordgo.ApplicationCommand{
		Name:                     "check",
		Description:              "High level summary about a player",
		Contexts:                 defaultCtx,
		DefaultMemberPermissions: &userPerms,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "steamid",
				Description: "SteamID/Profile URL",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	})
}

func onCheck(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) (*discordgo.MessageEmbed, error) {
	opts := OptionMap(interaction.ApplicationCommandData().Options)

	playerID, errPlayerID := steamid.Resolve(ctx, opts.String("steamid"))
	if errPlayerID != nil || !playerID.Valid() {
		return nil, steamid.ErrInvalidSID
	}

	return &discordgo.MessageEmbed{Description: "hi: " + playerID.String()}, nil
}

type CommandOptions map[string]*discordgo.ApplicationCommandInteractionDataOption

// OptionMap will take the recursive discord slash commands and flatten them into a simple
// map.
func OptionMap(options []*discordgo.ApplicationCommandInteractionDataOption) CommandOptions {
	optionM := make(CommandOptions, len(options))
	for _, opt := range options {
		optionM[opt.Name] = opt
	}

	return optionM
}

func (opts CommandOptions) String(key string) string {
	root, found := opts[key]
	if !found {
		return ""
	}

	val, ok := root.Value.(string)
	if !ok {
		return ""
	}

	return val
}
