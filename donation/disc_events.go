package donation

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func discEvents() {
	discord.MemberJoin.Add(func(s *discordgo.Session, v *discordgo.GuildMemberAdd) bool {
		v.Roles = setDonationRoles(s, c, v.User.ID, v.Roles)

		return false
	})

	discord.Ready.Add(func(s *discordgo.Session, v *discordgo.Ready) bool {
		emb := discutils.BaseEmbed
		emb.Title = "Donation Information"
		emb.Description = "There are rewards for donations. Some of the tiers receive a special donor-only channel here. All tiers also have a special XP Multiplier (see blow for table)\n**To receive these rewards, you must be logged when donating!**\nDonations are accepted [on my donation website](https://donate.shadygoat.eu)\nPlease also note that the lower bound is not included in the requirement! If it says 10-15, if one donates 10, they are not in this bracket, but if they donate 15 they are"
		roles := []*config.DonationRole{}
		roles = append(roles, *config.Donations.Monthly...)
		roles = append(roles, *config.Donations.Persistent...)

		for i, r := range roles {
			name1, name2, name3 := discutils.EMBED_EMPTY, discutils.EMBED_EMPTY, discutils.EMBED_EMPTY

			if i == 0 {
				name1, name2, name3 = "Role", "Requirement", "XP Multiplier"
			}

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name1,
				Value:  "<@&" + r.RoleID + ">",
				Inline: true,
			})

			req := fmt.Sprint(r.Min)

			if r.Min == 0 && r.Max == -1 {
				req = "Any Donation"
			} else {
				if r.Min == 0 {
					req = "<"
				} else if r.Max != -1 {
					req += "-"
				}
				if r.Max == -1 {
					req += "+"
				} else {
					req += fmt.Sprint(r.Max)
				}
			}

			if i < len(*config.Donations.Monthly) {
				req += " This Month"
			}

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name2,
				Value:  req,
				Inline: true,
			})

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name3,
				Value:  "x" + fmt.Sprint(r.XPMultiplier),
				Inline: true,
			})
		}

		if config.Donations.Info == "" {
			return false
		}

		msgs, err := s.ChannelMessages(config.Donations.Info, 1, "", "", "")
		shouldPurge := false

		if log.ErrorIfErr(err, "fetching donation info message") || len(msgs) != 1 || len(msgs[0].Embeds) != 1 || len(msgs[0].Embeds[0].Fields) != len(emb.Fields) {
			shouldPurge = true
		} else {
			oldEmb := msgs[0].Embeds[0]

			if oldEmb.Title != emb.Title || oldEmb.Description != emb.Description {
				shouldPurge = true
			} else {
				for i := range oldEmb.Fields {
					if oldEmb.Fields[i].Inline != emb.Fields[i].Inline || oldEmb.Fields[i].Name != emb.Fields[i].Name || oldEmb.Fields[i].Value != emb.Fields[i].Value {
						shouldPurge = true
						break
					}
				}
			}
		}

		if shouldPurge {
			discutils.PurgeChannel(s, config.Donations.Info)
			discutils.SendMessage(s, config.Donations.Info, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					&emb,
				},
			})
		}

		return false
	})
}
