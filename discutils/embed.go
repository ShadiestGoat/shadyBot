package discutils

import (
	"github.com/bwmarrin/discordgo"
)

const (
	EMBED_EMPTY = "\u200b"
	MY_PFP      = "https://media.discordapp.net/attachments/735886276070342696/1020345904164773918/pfp-f.png"

	COLOR_PRIMARY = 0xad6bee
	COLOR_SUCCESS = 0x08dd7e
	COLOR_DANGER  = 0xA51D2A
)

var BaseEmbed = discordgo.MessageEmbed{
	Title: EMBED_EMPTY,
	Footer: &discordgo.MessageEmbedFooter{
		Text:    "Made By Shady Goat",
		IconURL: "https://media.discordapp.net/attachments/735886276070342696/1020345904164773918/pfp-f.png",
	},
	Fields: []*discordgo.MessageEmbedField{},
	Color:  COLOR_PRIMARY,
}
