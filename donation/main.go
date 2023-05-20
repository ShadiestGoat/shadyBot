package donation

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/initializer"
)

var c *donations.Client

const fundNotFound = "Couldn't find the fund!\nYou [can view all the funds here](https://donate.shadygoat.eu/funds)"

var minAliasL = 3

func init() {
	cmdEditFund()
	cmdAddFund()
	cmdFund()
	cmdDonor()

	discEvents()

	discord.RegisterAutocomplete("fund", autocompleteFunds)
	discord.RegisterAutocomplete("editfund", autocompleteFunds)
	discord.RegisterAutocomplete("donate", autocompleteFunds)

	initializer.Register(initializer.MOD_DONATION, func(ctx *initializer.InitContext) {
		c = donations.NewClient(config.Donations.Token, donations.WithCustomLocation(config.Donations.Location))

		c.AddHandler(func(c *donations.Client, v *donations.EventOpen) {
			log.Debug("Reloading the donation WS conn")

			after := ""
			errors := 0

			for {
				if errors > 5 {
					log.Fatal("Got too many errors when initiating the donations")
				}

				members, err := ctx.Discord.GuildMembers(config.Discord.GuildID, after, 100)

				if log.ErrorIfErr(err, "fetching guild members in donation setup") {
					errors++
					time.Sleep(5 * time.Second)
				} else {

					for i, mem := range members {
						setDonationRoles(ctx.Discord, c, mem.User.ID, mem.Roles)
						if i == 99 {
							after = mem.User.ID
						}
					}

					if len(members) != 100 {
						break
					}

					errors = 0
				}
			}

			log.Debug("Finished the guild member donation setup")
		})

		c.AddHandler(func(c *donations.Client, v *donations.EventClose) {
			log.Warn("The donation API had to be shut down due to '%v', restarting in 30s...", v.Err)
			time.Sleep(30 * time.Second)
			c.OpenWS()
		})

		c.AddHandler(func(c *donations.Client, v *donations.EventNewDonation) {
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
				mem := discutils.GetMember(ctx.Discord, config.Discord.GuildID, donorDiscord)
				if mem != nil {
					setDonationRoles(ctx.Discord, c, discordID, mem.Roles)

					emb.Author = &discordgo.MessageEmbedAuthor{
						Name:    discutils.MemberName(mem),
						IconURL: mem.AvatarURL("128"),
					}
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

			if config.Donations.Donations != "" {
				discutils.SendMessage(ctx.Discord, config.Donations.Donations, &discordgo.MessageSend{
					Embeds: []*discordgo.MessageEmbed{
						&emb,
					},
				})
			}
		})

		c.AddHandler(func(c *donations.Client, v *donations.EventNewFund) {
			goalStr := ""

			if v.Goal != 0 {
				goalStr = " with a goal of " + fmt.Sprint(v.Goal) + " Euros"
			}
			emb := discutils.BaseEmbed
			emb.Title = "Fund '" + v.ShortTitle + "' has been created" + goalStr + "!"
			emb.Description = v.Title + "\n[You can look at it here](" + c.FundURL(v.Fund) + ")"
			discutils.SendMessage(ctx.Discord, config.Donations.Funds, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					&emb,
				},
			})
		})

		go donationQueue.Loop(ctx.Discord, c)
	}, &initializer.ModuleInfo{
		PreHooks: []initutils.Module{
		},
	}, initializer.MOD_DISCORD)
}
