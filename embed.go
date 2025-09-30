package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	avatarURLSmallFormat  = "https://avatars.akamai.steamstatic.com/%s.jpg"
	avatarURLMediumFormat = "https://avatars.akamai.steamstatic.com/%s_medium.jpg"
	avatarURLFullFormat   = "https://avatars.akamai.steamstatic.com/%s_full.jpg"
)

func NewAvatar(hash string) Avatar {
	return Avatar{hash: hash}
}

type Avatar struct {
	hash string
}

func (h Avatar) Full() string {
	return fmt.Sprintf(avatarURLFullFormat, h.hash)
}

func (h Avatar) Medium() string {
	return fmt.Sprintf(avatarURLMediumFormat, h.hash)
}

func (h Avatar) Small() string {
	return fmt.Sprintf(avatarURLSmallFormat, h.hash)
}

func (h Avatar) Hash() string {
	return h.hash
}

// func addField(embed *discordgo.MessageEmbed, name string, value string) {
// 	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
// 		Name:  name,
// 		Value: value,
// 	})
// }

func addFieldInline(embed *discordgo.MessageEmbed, name string, value string) {
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: true,
	})
}

func newEmbed(title string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		//URL: "https://steamcommunity.com/profiles/" + profile.SteamId,
		//Type: discordgo.EmbedTypeArticle,
		Title: title,
		// Thumbnail: &discordgo.MessageEmbedThumbnail{
		// 	URL:    NewAvatar(profile.AvatarHash).Medium(),
		// 	Width:  64,
		// 	Height: 64,
		// },
		Provider: &discordgo.MessageEmbedProvider{
			URL:  "https://tf-api.roto.lol",
			Name: "tf-api",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
