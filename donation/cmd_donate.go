package donation

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func init() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "donate",
		DefaultMemberPermissions: discord.Perms(),
		Description:              "Get auto logged in link for the donation website",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "fund",
				Description:  "The fund you want to donate to",
				Autocomplete: true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		fundID := "default"
		if data["fund"] != nil {
			fundID = data["fund"].StringValue()
		}

		fund, err := c.FundByID(fundID)
		if log.ErrorIfErr(err, "fetching fund '%s'", fundID) {
			discutils.IError(s, i.Interaction, fundNotFound)
			return
		}

		btn := discordgo.Button{
			Label: "Donate Now",
			Style: discordgo.LinkButton,
			URL:   c.FundURL(fund) + "?id=" + discutils.IAuthor(i.Interaction).ID,
		}

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: &discordgo.MessageEmbed{},
			Comps: []discordgo.MessageComponent{
				btn,
				btn,
				btn,
			},
		}, discutils.I_EPHEMERAL)
	})
}
