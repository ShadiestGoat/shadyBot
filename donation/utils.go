package donation

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func embedFund(f *donations.Fund) *discordgo.MessageEmbed {
	emb := discutils.BaseEmbed
	emb.Title = "The '" + f.ShortTitle + "' Fund"
	emb.Description = "[" + f.Title + "](" + c.FundURL(f) + ")"

	if f.Goal != 0 {
		if f.Amount == nil {
			tmpF, err := c.FundByID(f.ID)
			if log.ErrorIfErr(err, "fetching fund for embed <3") {
				return nil
			}

			*f = *tmpF
		}

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Goal",
			Value:  utils.TextProgressBar(f.Goal, *f.Amount, "0", fmt.Sprint(f.Goal), 36),
			Inline: false,
		})
	}

	return &emb
}

func sendDonationMessage(v *donations.Donation, s *discordgo.Session) (mem *discordgo.Member) {
	donor, err := c.DonorByID(v.Donor, false)
	if log.ErrorIfErr(err, "fetching donor '%s'", v.Donor) {
		return
	}

	fund, err := c.FundByID(v.FundID)
	if log.ErrorIfErr(err, "fetching fund '%s'", v.FundID) {
		return
	}

	donorDiscord := ""

	for _, d := range donor.Donors {
		if d.DiscordID != "" {
			donorDiscord = d.DiscordID
			break
		}
	}

	emb := discutils.BaseEmbed
	emb.Title = "New Donation!"

	if donorDiscord != "" {
		discordID := donorDiscord
		donorDiscord = "<@" + donorDiscord + ">"

		mem = discutils.GetMember(s, config.Discord.GuildID, discordID)

		emb.Author = &discordgo.MessageEmbedAuthor{
			Name:    discutils.MemberName(mem),
			IconURL: mem.AvatarURL("128"),
		}
	} else {
		donorDiscord = "Someone"
	}

	emb.Description = fmt.Sprintf("%s donated %.2f Euro for the [%s](%s) fund!", donorDiscord, v.Amount, fund.ShortTitle, c.FundURL(fund))
	emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
		Name:   "ID",
		Value:  "`" + v.ID + "`",
		Inline: false,
	})

	emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
		Name:   "Message",
		Value:  v.Message,
		Inline: false,
	})

	if config.Donations.ChanDonations != "" {
		discutils.SendMessage(s, config.Donations.ChanDonations, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				&emb,
			},
		})
	}

	return
}