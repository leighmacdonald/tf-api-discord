package main

import (
	"fmt"

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
