package donation

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func cmdFund() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "fund",
		DefaultMemberPermissions: discord.Perms(),
		Description:              "Fetch information about a fund",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "fund",
				Description:  "The fund to fetch info about",
				Required:     true,
				Autocomplete: true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		fundID := data["fund"].StringValue()
		fund, err := c.FundByID(fundID)
		if log.ErrorIfErr(err, "fetching fund '%s'", fundID) || fund == nil {
			discutils.IError(s, i.Interaction, "Couldn't find this fund :(")
			return
		}
		discutils.IEmbed(s, i.Interaction, embedFund(fund), discutils.I_EPHEMERAL)
	})
}
