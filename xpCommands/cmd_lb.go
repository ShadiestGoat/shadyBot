package xpCommands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
	"github.com/shadiestgoat/shadyBot/xp"
)

func cmdLeaderboard() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "leaderboard",
		Description: "view xp leaderboards",
		Options: []*discordgo.ApplicationCommandOption{
			orderOpt(false),
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "hidden",
				Description: "Make the response appear only to you (default: true)",
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		hide := true
		if opt := data["hidden"]; opt != nil {
			hide = opt.BoolValue()
		}

		order := ORDER_XP
		if opt := data["type"]; opt != nil {
			order = opt.StringValue()
		}

		respOpts := discutils.I_NONE
		if hide {
			respOpts |= discutils.I_EPHEMERAL
		}

		emb, err := leaderboardEmbed(order, 1)

		if log.ErrorIfErr(err, "creating lb embed for order '%s' page '%d'", order, 1) {
			discutils.IError(s, i.Interaction, "Couldn't fetch leaderboard :(")
			return
		}

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: emb,
			Comps: buttonFactory(order, 1, false),
		}, respOpts)
	})

	discord.RegisterComponent("xpLB", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		info, newPage := utils.ParsePagination(d.CustomID, utils.SUPER_NUM)
		order := info[0]

		emb, err := leaderboardEmbed(order, newPage)

		if log.ErrorIfErr(err, "creating lb embed for order (btn) '%s' page '%d'", order, newPage) {
			discutils.IError(s, i.Interaction, "Couldn't fetch leaderboard :(", discutils.I_UPDATE)
			return
		}

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: emb,
			Comps: buttonFactory(order, newPage, false),
		}, discutils.I_UPDATE)
	})
}

func maxSize(order string) int {
	size := 0
	db.QueryRow(`SELECT COUNT(*) FROM xp`, nil, &size)

	return size
}

func buttonFactory(order string, curPage int, fullDisable bool) []discordgo.MessageComponent {
	return utils.PaginationButtonFactory(curPage, maxSize(order), "xpLB", fullDisable, utils.SUPER_NUM, order)
}

func leaderboardEmbed(order string, page int) (*discordgo.MessageEmbed, error) {
	emb := discutils.BaseEmbed

	switch order {
	case ORDER_MSG:
		emb.Title = "Writer"
	case ORDER_VC:
		emb.Title = "Speaker"
	case ORDER_XP:
		emb.Title = "XP"
	}

	emb.Title += " Leaderboard"

	start, stop := utils.PageBounds(page)
	info, err := FetchRows(order, start, stop)
	if err != nil {
		return nil, err
	}

	fieldNames := [3]string{
		"",
		"Stats",
		discutils.EMBED_EMPTY,
	}

	switch order {
	case ORDER_MSG:
		fieldNames[0] = "Writer"
	case ORDER_VC:
		fieldNames[0] = "Speaker"
	case ORDER_XP:
		fieldNames[0] = "User"
		fieldNames[2] = "Progress"
	}

	for i, row := range info {
		f := []*discordgo.MessageEmbedField{
			{
				Name:   discutils.EMBED_EMPTY,
				Value:  discutils.EMBED_EMPTY,
				Inline: true,
			},
			{
				Name:   discutils.EMBED_EMPTY,
				Value:  discutils.EMBED_EMPTY,
				Inline: true,
			},
			{
				Name:   discutils.EMBED_EMPTY,
				Value:  discutils.EMBED_EMPTY,
				Inline: true,
			},
		}

		if i == 0 {
			for j := 0; j < 3; j++ {
				f[j].Name = fieldNames[j]
			}
		}

		rank := 0

		msgs := "**" + utils.FormatShortInt(row.Messages) + "** Messages"
		vc := "**" + utils.FormatMinutes(row.VoiceMinutes) + "** in VC"

		switch order {
		case ORDER_MSG:
			rank = RankMsg(row.Messages, row.LastActive)
			f[1].Value = msgs + "\n" + vc
		case ORDER_VC:
			rank = RankVC(row.VoiceMinutes, row.LastActive)
			f[1].Value = vc + "\n" + msgs
		case ORDER_XP:
			rank = RankXP(row.Level, row.XP, row.LastActive)
			lvlReq := xp.LevelUpRequirement(row.Level)

			f[1].Value = msgs + "\n" + vc
			f[2].Value = utils.TextProgressBar(float64(lvlReq), float64(row.XP), fmt.Sprint(row.Level), fmt.Sprint(row.Level+1), 12)
		}

		f[0].Value = fmt.Sprintf("%d. <@%s>", rank, row.UserID)

		emb.Fields = append(emb.Fields, f...)
	}

	return &emb, nil
}
