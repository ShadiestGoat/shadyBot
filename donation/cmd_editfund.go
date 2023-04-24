package donation

import (
	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func cmdEditFund() {
	minFundL := 3
	minGoal := 10.0

	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "editfund",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Description:              "Edit a fund by id!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "fund",
				Description:  "The ID of the fund",
				Required:     true,
				Autocomplete: true,
				MinLength:    &minFundL,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "short-title",
				Description: "The short title - The 'XX' fund",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "The longer title used as the big text on the main page",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "alias",
				Description: "The alias of the fund - https:///f/XX",
				Required:    false,
				MaxLength:   16,
				MinLength:   &minAliasL,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "default",
				Description: "Whether to make this fund the default one on the main page",
			},
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "goal",
				Description: "Include a value if you want this fund to have a donation goal",
				MinValue:    &minGoal,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		fundID := data["fund"].StringValue()
		fund, err := c.FundByID(fundID)

		if err != nil {
			errMsg := ""

			if err.(*donations.HTTPError).Status == 404 {
				errMsg = fundNotFound
			} else {
				log.ErrorIfErr(err, "fetching fund '%s'", fundID)
				errMsg = "Unknown error when fetching funds!"
			}

			discutils.IError(s, i.Interaction, errMsg)

			return
		}

		fund.Amount = nil

		ogDefault := *fund.Default

		rels := map[string]any{
			"short-title": &fund.ShortTitle,
			"title":       &fund.Title,
			"alias":       &fund.Alias,
			"default":     &fund.Default,
			"goal":        &fund.Goal,
		}

		for opt, r := range rels {
			cmd := data[opt]
			if cmd == nil {
				continue
			}
			switch cmd.Type {
			case discordgo.ApplicationCommandOptionString:
				*(r.(*string)) = cmd.StringValue()
			case discordgo.ApplicationCommandOptionNumber:
				*(r.(*float64)) = cmd.FloatValue()
			case discordgo.ApplicationCommandOptionBoolean:
				tmp := cmd.BoolValue()
				*(r.(**bool)) = &tmp
			}
		}

		err = c.UpdateFund(fund)
		if log.ErrorIfErr(err, "updating fund '%s'", fundID) {
			discutils.IError(s, i.Interaction, "Couldn't update fund :(")
			return
		}

		if ogDefault != *fund.Default && *fund.Default {
			_, err := c.MakeFundDefault(fundID)

			if log.ErrorIfErr(err, "updating fund '%s'", fundID) {
				discutils.IError(s, i.Interaction, "Couldn't make the fund default!")
				return
			}
		}

		discutils.IEmbed(s, i.Interaction, embedFund(fund), discutils.I_EPHEMERAL)
	})
}
