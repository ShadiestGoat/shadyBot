package discutils

import (
	"github.com/bwmarrin/discordgo"
)

func MemberName(mem *discordgo.Member) string {
	if mem.Nick == "" {
		return mem.User.Username
	}
	return mem.Nick
}

func IAuthor(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}

	return i.User
}
