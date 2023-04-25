package xp

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func multiplier(s *discordgo.Session, userID string) float64 {
	mem := discutils.GetMember(s, config.Discord.GuildID, userID)
	if mem == nil {
		return 1.0
	}

	return multiplierDonation(mem) * multiplierOwner(mem) * multiplierPremium(mem)
}

func multiplierPremium(mem *discordgo.Member) float64 {
	if discutils.IsMemberPremium(mem) {
		return 1.3
	}
	return 1
}

// Done since I can't be a donor :(
func multiplierOwner(mem *discordgo.Member) float64 {
	if mem.User.ID == config.OWNER_ID {
		return 1.75
	}
	return 1
}

func multiplierDonation(mem *discordgo.Member) float64 {
	mult := 1.0

	m := map[string]float64{}
	allRoles := []*config.DonationRole{}
	allRoles = append(allRoles, *config.Donations.Persistent...)
	allRoles = append(allRoles, *config.Donations.Monthly...)
	for _, r := range allRoles {
		m[r.RoleID] = r.XPMultiplier
	}

	for _, r := range mem.Roles {
		if m[r] > mult {
			mult = m[r]
		}
	}

	return mult
}
