package xpCommands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
	"github.com/shadiestgoat/shadyBot/xp"
)

// /rank
// - leaderboard: Voice Time | Message Number | Experience
// - user
// - rank: query based on the user's rank
// - hidden
// {
// 	Type:        discordgo.ApplicationCommandOptionBoolean,
// 	Name:        "hidden",
// 	Description: "Make the response appear only to you (default: true)",
// 	Required:    false,
// },
// {
// 	Type:        discordgo.ApplicationCommandOptionUser,
// 	Name:        "user",
// 	Description: "Query a user's rank, takes priority over rank",
// 	Required:    false,
// },
// {
// 	Type:        discordgo.ApplicationCommandOptionInteger,
// 	Name:        "rank",
// 	Description: "Query based on a person's rank",
// 	Required:    false,
// }

var hiddenOpt = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionBoolean,
	Name:        "hidden",
	Description: "Make the response appear only to you (default: true)",
	Required:    false,
}

func cmdRank() {
	minRank := 1.0

	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "rank",
		Description: "view the rank of someone in the leaderboards",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "user",
				Description: "Fetch the rank of a specific user (yourself by default)",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "The user to fetch from (default: you)",
						Required:    false,
					},
					hiddenOpt,
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "query",
				Description: "Fetch the rank based on a query",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "rank",
						Description: "The rank you are asking for",
						Required:    true,
						MinValue:    &minRank,
					},
					orderOpt(false),
					hiddenOpt,
				},
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		cmd1 := data["cmd-1"].StringValue()

		hidden := true

		if data["hidden"] != nil {
			hidden = data["hidden"].BoolValue()
		}

		rank := &Rank{}

		switch cmd1 {
		case "user":
			userID := discutils.IAuthor(i.Interaction).ID

			if data["user"] != nil {
				userID = data["user"].UserValue(nil).ID
			}

			rank = FetchUser(userID)
		case "query":
			order := ORDER_XP

			if data["type"] != nil {
				order = data["type"].StringValue()
			}

			rank = FetchRank(order, int(data["rank"].IntValue()))
		}
		if rank == nil {
			discutils.IError(s, i.Interaction, "Couldn't find this XP User!")
			return
		}

		rank.FillRanks()

		respOpts := discutils.I_NONE

		if hidden {
			respOpts |= discutils.I_EPHEMERAL
		}

		emb := discutils.BaseEmbed

		name := ""
		pfp := ""
		id := ""

		if i.Member != nil {
			name = discutils.MemberName(i.Member)
			pfp = i.Member.AvatarURL("256")
			id = i.Member.User.ID
		} else {
			name = i.User.Username
			pfp = i.User.AvatarURL("256")
			id = i.User.ID
		}

		emb.Title = name + "'s Rankings"
		emb.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: pfp,
		}
		emb.Author = &discordgo.MessageEmbedAuthor{
			Name:    name,
			IconURL: pfp,
		}

		emb.Description = "Here are <@" + id + ">'s rankings across all the categories!"

		reqLvl := xp.LevelUpRequirement(rank.Level)

		emb.Fields = append(emb.Fields, []*discordgo.MessageEmbedField{
			{
				Name:   "Messages",
				Value:  "Rank **#" + fmt.Sprint(rank.RankMsg) + "** " + utils.FormatShortInt(rank.Messages),
				Inline: true,
			},
			{
				Name:   "VC Time",
				Value:  "Rank **#" + fmt.Sprint(rank.RankVC) + "** " + utils.FormatMinutes(rank.VoiceMinutes),
				Inline: true,
			},
			{
				Name: "Experience",
				Value: "Rank **#" + fmt.Sprint(rank.RankXP) + "**\n" + utils.TextProgressBar(
					float64(reqLvl), float64(rank.XP),
					fmt.Sprint(rank.Level), fmt.Sprint(rank.Level+1),
					36,
				),
				Inline: false,
			},
		}...)

		discutils.IEmbed(s, i.Interaction, &emb, respOpts)
	})
}
