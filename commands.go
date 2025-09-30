package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leighmacdonald/discordgo-lipstick/bot"
	"github.com/leighmacdonald/steamid/v4/steamid"
	"github.com/leighmacdonald/tf-api-discord/tfapi"
)

func registerCommands(ctx context.Context, discord *bot.Bot, api *tfapi.TFAPI) error {
	sites, err := api.MetaSitesWithResponse(ctx)
	if err != nil {
		return err
	}

	var siteNames []*discordgo.ApplicationCommandOptionChoice
	for _, site := range *sites.JSON200 {
		siteNames = append(siteNames, &discordgo.ApplicationCommandOptionChoice{
			Name:  site.Title,
			Value: site.Name,
		})
	}

	defaultCtx := &[]discordgo.InteractionContextType{discordgo.InteractionContextBotDM}
	//modPerms := int(discordgo.PermissionBanMembers)
	userPerms := int64(discordgo.PermissionViewChannel)

	discord.MustRegisterHandler("check", &discordgo.ApplicationCommand{
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
	}, onCheck(api))

	discord.MustRegisterHandler("bans", &discordgo.ApplicationCommand{
		Name:                     "bans",
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
			{
				Name:        "site",
				Description: "Limit results to a specific site",
				Type:        discordgo.ApplicationCommandOptionString,
				Choices:     siteNames,
				Required:    false,
			},
		},
	}, onBans(api))

	discord.MustRegisterHandler("stats", &discordgo.ApplicationCommand{
		Name:                     "stats",
		Description:              "Stats about the underlying database",
		Contexts:                 defaultCtx,
		DefaultMemberPermissions: &userPerms,
	}, onStats(api))

	return nil
}

func onStats(api *tfapi.TFAPI) bot.Handler {
	return func(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) (*discordgo.MessageEmbed, error) {
		resp, errResp := api.StatsIdWithResponse(ctx)
		if errResp != nil {
			return nil, errResp
		}

		stats := *resp.JSON200

		embed := &discordgo.MessageEmbed{
			//Type: discordgo.EmbedTypeArticle,
			Title: "[Stats] Overall",
			Provider: &discordgo.MessageEmbedProvider{
				URL:  "https://tf-api.roto.lol",
				Name: "tf-api",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		addFieldInline(embed, "Ban Total Count", strconv.Itoa(int(stats.BanTotalCount)))
		addFieldInline(embed, "Bot Detector Lists", strconv.Itoa(int(stats.BdListCount)))
		addFieldInline(embed, "Bot Detector Entries", strconv.Itoa(int(stats.BdListEntriesCount)))

		addFieldInline(embed, "Vac Counts", strconv.Itoa(int(stats.VacCount)))
		addFieldInline(embed, "Game Ban Counts", strconv.Itoa(int(stats.GameBanCount)))
		addFieldInline(embed, "Comm Ban Counts", strconv.Itoa(int(stats.CommunityBanCount)))

		addFieldInline(embed, "LogsTF Logs", strconv.Itoa(int(stats.LogsTfCount)))
		addFieldInline(embed, "LogsTF Players", strconv.Itoa(int(stats.LogsTfPlayerCount)))
		addFieldInline(embed, "LogsTF Messages", strconv.Itoa(int(stats.LogsTfChatCount)))

		addFieldInline(embed, "Sources (Sourcebans)", strconv.Itoa(int(stats.SourceCount)))
		addFieldInline(embed, "Sources (Leagues)", strconv.Itoa(int(stats.LeaguesCount)))
		addFieldInline(embed, "League Teams", strconv.Itoa(int(stats.LeaguesTeamCount)))

		addFieldInline(embed, "Names", strconv.Itoa(int(stats.NameCount)))
		addFieldInline(embed, "Avatars", strconv.Itoa(int(stats.AvatarCount)))
		addFieldInline(embed, "Friends", strconv.Itoa(int(stats.FriendCount)))

		return embed, nil
	}
}

func onCheck(api *tfapi.TFAPI) bot.Handler {
	return func(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) (*discordgo.MessageEmbed, error) {
		opts := bot.OptionMap(interaction.ApplicationCommandData().Options)

		playerID, errPlayerID := steamid.Resolve(ctx, opts.String("steamid"))
		if errPlayerID != nil || !playerID.Valid() {
			return nil, steamid.ErrInvalidSID
		}

		resp, errResp := api.MetaProfileWithResponse(ctx, &tfapi.MetaProfileParams{
			Steamids: playerID.String(),
		})
		if errResp != nil {
			return nil, errResp
		}

		profiles := *resp.JSON200
		if len(profiles) != 1 {
			return nil, fmt.Errorf("%w: Invalid response count", bot.ErrCommandExec)
		}
		profile := profiles[0]

		embed := &discordgo.MessageEmbed{
			URL: "https://steamcommunity.com/profiles/" + profile.SteamId,
			//Type: discordgo.EmbedTypeArticle,
			Title: "[Check] " + profile.PersonaName,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    NewAvatar(profile.AvatarHash).Medium(),
				Width:  64,
				Height: 64,
			},
			Provider: &discordgo.MessageEmbedProvider{
				URL:  "https://tf-api.roto.lol",
				Name: "tf-api",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		addFieldInline(embed, "SteamID", profile.SteamId)
		addFieldInline(embed, "Name", profile.PersonaName)
		addFieldInline(embed, "Real Name", profile.RealName)
		addFieldInline(embed, "Account Created", time.Unix(profile.TimeCreated, 0).Format(time.DateOnly))
		addFieldInline(embed, "Community Ban", strconv.FormatBool(profile.CommunityBanned))
		addFieldInline(embed, "Econ Ban", profile.EconomyBan)
		addFieldInline(embed, "Vac Bans", strconv.Itoa(int(profile.NumberOfVacBans)))
		addFieldInline(embed, "Sourcebans", strconv.Itoa(len(profile.Bans)))
		addFieldInline(embed, "Comp Teams", strconv.Itoa(len(profile.CompetitiveTeams)))

		return embed, nil
	}
}

func onBans(api *tfapi.TFAPI) bot.Handler {
	return func(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) (*discordgo.MessageEmbed, error) {
		opts := bot.OptionMap(interaction.ApplicationCommandData().Options)

		playerID, errPlayerID := steamid.Resolve(ctx, opts.String("steamid"))
		if errPlayerID != nil || !playerID.Valid() {
			return nil, steamid.ErrInvalidSID
		}

		resp, errResp := api.BansSearchWithResponse(ctx, &tfapi.BansSearchParams{Steamids: playerID.String()})
		if errResp != nil {
			return nil, errResp
		}

		bans := *resp.JSON200
		if len(bans) != 11 {
			return nil, fmt.Errorf("%w: Invalid response count", bot.ErrCommandExec)
		}

		embed := newEmbed("[Bans] History")
		embed.URL = "https://steamcommunity.com/profiles/" + playerID.String()

		return embed, nil
	}
}
