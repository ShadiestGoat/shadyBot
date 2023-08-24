package donation

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func cmdAddFund() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Name:                     "addfund",
		Description:              "Create a new fund on the donation website",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "short-title",
				Description: "The short title - The 'XX' fund",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "The longer title used as the big text on the display page",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "alias",
				Description: "The alias of the fund - /f/XX",
				Required:    true,
				MaxLength:   16,
				MinLength:   &minAliasL,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "default",
				Description: "Whether to make this fund the default one on the main page (default: false)",
			},
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "goal",
				Description: "Include a value if you want this fund to have a donation goal",
				// Must be 0 or above!
				MinValue: new(float64),
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		isDefault := false
		if data["default"] != nil {
			isDefault = data["default"].BoolValue()
		}

		goal := 0.0
		if data["goal"] != nil {
			goal = data["goal"].FloatValue()
		}

		fund, err := c.NewFund(data["alias"].StringValue(), data["short-title"].StringValue(), data["title"].StringValue(), isDefault, goal)

		if err != nil {
			discutils.IError(s, i.Interaction, err.Error())
			return
		}

		if fund.Amount == nil {
			fund.Amount = new(float64)
		}

		emb := embedFund(fund)
		discutils.IEmbed(s, i.Interaction, emb, discutils.I_EPHEMERAL)
	})
}
