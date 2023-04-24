package donation

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func cmdDonor() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.ChatApplicationCommand,
		Name:              "donor",
		DefaultMemberPermissions: discord.Perms(),
		Description:       "Fetch information about a donor",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "donor",
				Description: "The donor to fetch info about",
				Required:    true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		userID := data["donor"].UserValue(nil).ID
		resp, err := c.DonorByDiscord(userID, true)
		if err != nil && err.(*donations.HTTPError).Status != 404 {
			if log.ErrorIfErr(err, "fetching discord donor '%s'", userID) {
				discutils.IError(s, i.Interaction, "Unknown Error!")
			}
		}

		if err != nil || len(resp.Donors) == 0 {
			discutils.IError(s, i.Interaction, "Not a registered donor!")
			return
		}

		IDs := ""
		for _, d := range resp.Donors {
			IDs += "`" + d.ID + "`\n"
		}
		emb := discutils.BaseEmbed
		emb.Title = "Donor Profile"
		emb.Description = "This is the the donor profile of <@" + userID + ">"
		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Donor IDs",
			Value:  IDs,
			Inline: true,
		})
		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Donated this month",
			Value:  fmt.Sprint(resp.Total.Month),
			Inline: true,
		})
		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Donated total",
			Value:  fmt.Sprint(resp.Total.Total),
			Inline: true,
		})

		donos := *resp.Donations
		if len(donos) > 7 {
			donos = donos[:7]
		}

		funds := map[string]*donations.Fund{}

		for _, d := range donos {
			if funds[d.FundID] == nil {
				funds[d.FundID], err = c.FundByID(d.FundID)
				if log.ErrorIfErr(err, "fetching fund '%s'", d.FundID) {
					discutils.IError(s, i.Interaction, "Couldn't fetch funds :(")
				}
			}
		}

		for i, d := range donos {
			name1, name2, name3 := discutils.EMBED_EMPTY, discutils.EMBED_EMPTY, discutils.EMBED_EMPTY

			if i == 0 {
				name1, name2, name3 = "Amount Donated", "Fund", "Message"
			}

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name1,
				Value:  "€" + fmt.Sprint(d.Amount),
				Inline: true,
			})

			fund := funds[d.FundID]

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name2,
				Value:  fmt.Sprintf("[%v](%v)", fund.ShortTitle, c.FundURL(fund)),
				Inline: true,
			})

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name3,
				Value:  d.Message,
				Inline: true,
			})
		}

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_NONE)
	})
}
