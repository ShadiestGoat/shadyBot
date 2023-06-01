package donation

import (
	"fmt"
	"regexp"
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
	initializer.Register(initializer.MOD_DONATION_LOAD, func(c *initializer.InitContext) {
		cmdEditFund()
		cmdAddFund()
		cmdFund()
		cmdDonor()
	
		discEvents()
	
		discord.RegisterAutocomplete("fund", autocompleteFunds)
		discord.RegisterAutocomplete("editfund", autocompleteFunds)
		discord.RegisterAutocomplete("donate", autocompleteFunds)
	}, &initializer.ModuleInfo{
		PreHooks: []initutils.Module{initializer.MOD_DISCORD},
	})

	topicLocation := [2]int{}

	initializer.Register(initializer.MOD_DONATION, func(ctx *initializer.InitContext) {
		c = donations.NewClient(config.Donations.Token, donations.WithCustomLocation(config.Donations.Location))

		c.AddHandler(func(c *donations.Client, v *donations.EventOpen) {
			log.Debug("Opened DonationAPI WS")

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

			donoChan := discutils.GetChannel(ctx.Discord, config.Donations.ChanDonations)

			lastID := ""

			// New donations end up here <3 (last donation: {{id}})
			if len(donoChan.Topic) < (len(config.Donations.ChannelTopic) + (18-len("{{id}}"))) {
				log.Warn("Current donation channel topic not inline with the needed topic. Assuming no backlog...")
			} else {
				// {{id}} must be present since this thing hasn't crashed yet
				loc := regexp.MustCompile(`{{id}}`).FindStringIndex(config.Donations.ChannelTopic)

				topicLocation = [2]int{loc[0], len(config.Donations.ChannelTopic)-loc[1]}
				
				lastID = donoChan.Topic[topicLocation[0]:len(donoChan.Topic)-topicLocation[1]]
			}

			errCount := 0

			for {
				log.Debug("Trying for backlog...")

				donations, err := c.Donations("", lastID)
				
				if err != nil {
					if errCount > 4 {
						log.Fatal("Can't fetch donation backlog: %v", err)
					}

					time.Sleep(10 * time.Second)
					continue
				}

				if len(donations) == 0 {
					break
				}
				
				for _, d := range donations {
					sendDonationMessage(d, ctx.Discord)
				}

				lastID = donations[len(donations)-1].ID

				if len(donations) != 50 {
					break
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
			mem := sendDonationMessage(v.Donation, ctx.Discord)

			if mem != nil {
				setDonationRoles(ctx.Discord, c, mem.User.ID, mem.Roles)
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
			discutils.SendMessage(ctx.Discord, config.Donations.ChanFunds, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					&emb,
				},
			})
		})

		go donationQueue.Loop(ctx.Discord, c)
	}, &initializer.ModuleInfo{
		PreHooks: []initutils.Module{},
	}, initializer.MOD_DISCORD)
}
