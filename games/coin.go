package games

import (
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func cmdCoin() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "coinflip",
		Description: "Flip a coin! You can bet xp or not",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "side",
				Description: "The side you are betting on",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Heads",
						Value: 0,
					},
					{
						Name:  "Tails",
						Value: 1,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "bet",
				Description: "The XP you are betting on this flip",
				MinValue:    &config.BET_MIN,
				MaxValue:    config.BET_MAX,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		bet := 0
		// The side that the user predicted, 0 - heads, 1 - tails
		predictedSide := -1

		betRaw, sideRaw := data["bet"], data["side"]
		if betRaw != nil {
			bet = int(betRaw.IntValue())
		}
		if sideRaw != nil {
			predictedSide = int(sideRaw.IntValue())
		}

		if bet != 0 {
			if InActivityErrorCheck(s, i.Interaction) {
				return
			}
			if XPErrorCheck(s, i.Interaction, bet) {
				return
			}
			ActivityStore.Add(i.Interaction, GT_COIN)
		}

		realSide := utils.RandInt(0, 1)
		emb := discutils.BaseEmbed
		emb.Title = TITLE_COIN

		emb.Description = "Heads!"
		if realSide == 1 {
			emb.Description = "Tails!"
		}

		if predictedSide != -1 {
			if predictedSide == realSide {
				emb.Description += " You got it **right**!"
			} else {
				emb.Color = discutils.COLOR_DANGER
				emb.Description += " Ha! You got **L+Ratioed** by a coin"
			}
		}

		if bet == 0 {
			discutils.IEmbed(s, i.Interaction, &emb, discutils.I_NONE)
			return
		}

		affectedXP := bet

		if predictedSide != realSide {
			affectedXP *= -1
			emb.Description += "\nYou **lost** " + fmt.Sprint(bet) + "xp"
		} else {
			affectedXP = int(math.Round(float64(bet) * 0.9))
			emb.Description += "\nYou **won** 90% of your bet, ie. you gained **" + fmt.Sprint(affectedXP) + "xp**"
		}

		authorID := discutils.IAuthor(i.Interaction).ID
		FinishGame(authorID, affectedXP, affectedXP > 0, GT_COIN)

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_NONE)
	})
}
