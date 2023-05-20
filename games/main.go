package games

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/games/connect4"
	"github.com/shadiestgoat/shadyBot/initializer"
	"github.com/shadiestgoat/shadyBot/utils"
)

func init() {
	closer := make(chan bool, 2)

	disabledAllGambling := false

	initializer.Register(initializer.MOD_GAMES, func(c *initializer.InitContext) {
		disabled := config.Games.Disable

		if !disabled["connect4"] {
			connect4.Init()
		}

		if !disabled["blackjack"] {
			cmdBlackjack()
		}

		if !disabled["slots"] {
			cmdSlots()
		}
		
		if !disabled["coinflip"] {
			cmdCoin()
		}

		disabledAllGambling = disabled["slots"] && disabled["coinflip"] && disabled["blackjack"]

		if !disabledAllGambling {
			cmdGambler()
		}
	}, &initializer.ModuleInfo{
		PreHooks: []initutils.Module{
			initializer.MOD_DISCORD,
		},
	})
	
	initializer.Register(initializer.MOD_GAMBLER, func(c *initializer.InitContext) {
		if !disabledAllGambling {
			go ActivityStore.Loop(c.Discord, closer)
		}
	}, nil, initializer.MOD_DISCORD)

	initializer.RegisterCloser(initializer.MOD_GAMES, func() {
		closer <- true
	})
}

func cmdGambler() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "gambler",
		DefaultMemberPermissions: discord.Perms(),
		Description:              "Fetch information about a gambler",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The gambler to fetch info about",
				Required:    false,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		var userID string

		if data["user"] != nil {
			userID = data["user"].UserValue(nil).ID
		} else {
			userID = discutils.IAuthor(i.Interaction).ID
		}

		emb := discutils.BaseEmbed

		emb.Title = "Gambler Info"
		emb.Description = "<@" + userID + ">'s gambler profile"

		// game: [won, lost, gcd]
		m := map[GameType][3]int{}
		rows, err := db.Query(`SELECT game, won, lost FROM gambling WHERE id = $1`, userID)
		if err != nil {
			discutils.IError(s, i.Interaction, "Couldn't fetch your stats!")
			return
		}

		total := [2]int{}

		for rows.Next() {
			gt, won, lost := GT_BJ, 0, 0
			err = rows.Scan(&gt, &won, &lost)
			if log.ErrorIfErr(err, "scanning a gambler fetch") {
				discutils.IError(s, i.Interaction, "Couldn't fetch your stats!")
				return
			}

			m[gt] = [3]int{won, lost, utils.GreatestCommonDivisor(won, lost)}

			total[0] += won
			total[1] += lost
		}

		gTypes := []GameType{GT_BJ, GT_COIN, GT_SLOTS}

		for i, g := range gTypes {
			name1, name2, name3 := discutils.EMBED_EMPTY, discutils.EMBED_EMPTY, discutils.EMBED_EMPTY

			if i == 0 {
				name1, name2, name3 = "Game", "Won:Lost Ratio", "Total"
			}

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name1,
				Value:  g.String(),
				Inline: true,
			})

			if info, ok := m[g]; ok {
				emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
					Name:   name2,
					Value:  fmt.Sprint(info[0]/info[2]) + ":" + fmt.Sprint(info[1]/info[2]),
					Inline: true,
				})

				emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
					Name:   name3,
					Value:  fmt.Sprint(info[0] - info[1]),
					Inline: true,
				})
			} else {
				emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
					Name:   name2,
					Value:  "<no data>",
					Inline: true,
				})
				emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
					Name:   name3,
					Value:  "<no data>",
					Inline: true,
				})
			}
		}

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  "**Total**",
			Inline: true,
		})

		finGCD := utils.GreatestCommonDivisor(total[0], total[1])

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  fmt.Sprint(total[0]/finGCD) + ":" + fmt.Sprint(total[1]/finGCD),
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  fmt.Sprint(total[0] - total[1]),
			Inline: true,
		})

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_NONE)
	})
}
