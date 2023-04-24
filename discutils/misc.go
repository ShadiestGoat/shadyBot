package discutils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/config"
)

const (
	EMOJI_CROSS = "❌"
	EMOJI_TICK  = "✅"
)

func IsMemberPremium(mem *discordgo.Member) bool {
	return mem != nil && mem.PremiumSince != nil && !mem.PremiumSince.IsZero()
}

func HasRole(m *discordgo.Member, roleNeeded string) bool {
	for _, r := range m.Roles {
		if r == roleNeeded {
			return true
		}
	}
	return false
}

func EMessageURL(msg *discordgo.Message) string {
	return MessageURL(msg.ChannelID, msg.ID)
}

func MessageURL(Channel, Message string) string {
	return "https://discord.com/channels/" + config.Discord.GuildID + "/" + Channel + "/" + Message
}
