package warnings

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/snownode"
	"github.com/shadiestgoat/shadyBot/utils"
)

const super_num = utils.SUPER_NUM_SMALL

func formatWithSeverity(amt, severity int) string {
	str := fmt.Sprint(amt)

	if amt != severity {
		str += ", severity being " + fmt.Sprint(severity)
	}

	return str
}

func warningsEmbedNoUser(page int) (*discordgo.MessageEmbed, error) {
	start, stop := utils.PageBounds(page)

	emb := discutils.BaseEmbed

	emb.Title = "All Warnings"

	rows, err := db.Query(`SELECT id, warned_user, severity, reason FROM warnings ORDER BY id DESC LIMIT $1 OFFSET $2`, stop-start, start)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		id, userID, severity, reason := "", "", 0, ""

		err := rows.Scan(&id, &userID, &severity, &reason)

		if err != nil {
			return nil, err
		}

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  snownode.SnowToTime(id).Format("02/01/06 15:04") + "\n" + id + "\nSeverity: " + fmt.Sprint(severity),
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  "<@" + userID + ">",
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  reason,
			Inline: true,
		})
	}

	if len(emb.Fields) >= 3 {
		emb.Fields[0].Name = "Info"
		emb.Fields[1].Name = "User"
		emb.Fields[2].Name = "Reason"
	}

	return &emb, nil
}

func warningsEmbed(userID string, page int) (*discordgo.MessageEmbed, error) {
	if userID == "" {
		return warningsEmbedNoUser(page)
	}

	amountTot, amountTmp := Amount(userID)
	severityTot, severityTmp := Severity(userID)

	emb := discutils.BaseEmbed
	emb.Title = "Warnings"
	emb.Description = "These are <@" + userID + ">'s warnings\nTotal: " + formatWithSeverity(amountTot, severityTot) + "\nTemporary: " + formatWithSeverity(amountTmp, severityTmp)

	start, stop := utils.PageBounds(page)

	rows, err := db.Query(`SELECT id, severity, reason FROM warnings WHERE warned_user = $3 ORDER BY id DESC LIMIT $1 OFFSET $2`, stop-start, start, userID)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		id, severity, reason := "", 0, ""

		err := rows.Scan(&id, &severity, &reason)

		if err != nil {
			return nil, err
		}

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  snownode.SnowToTime(id).Format("02/01/06 15:04") + "\n" + id,
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  fmt.Sprint(severity),
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  reason,
			Inline: true,
		})
	}

	if len(emb.Fields) >= 3 {
		emb.Fields[0].Name = "Time/ID"
		emb.Fields[1].Name = "Severity"
		emb.Fields[2].Name = "Reason"
	}

	return &emb, nil
}

func buttonFactory(userID string, curPage int, fullDisable bool) []discordgo.MessageComponent {
	maxSize := 0

	qSuffix := ""
	conditions := []any{}

	if userID != "" {
		qSuffix = " WHERE warned_user = $1"
		conditions = append(conditions, userID)
	}

	db.QueryRow(`SELECT COUNT(*) FROM warnings`+qSuffix, conditions, &maxSize)

	return utils.PaginationButtonFactory(curPage, maxSize, "warnings", fullDisable, super_num, userID)
}

func cmdWarnings() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "warnings",
		Description: "View a list of your warnings",
		Options:     []*discordgo.ApplicationCommandOption{},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		authorID := discutils.IAuthor(i.Interaction).ID

		emb, err := warningsEmbed(authorID, 1)

		if log.ErrorIfErr(err, "creating a warnings embed for user '%s' page '1", authorID) {
			discutils.IError(s, i.Interaction, "Couldn't fetch the warnings :(")
			return
		}

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: emb,
			Comps: buttonFactory(authorID, 1, false),
		}, discutils.I_EPHEMERAL)
	})

	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "admin-warnings",
		Description: "View a list a warnings (admin variant)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to query by (default = no user)",
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		userID := ""

		if data["user"] != nil {
			userID = data["user"].UserValue(nil).ID
		}

		emb, err := warningsEmbed(userID, 1)

		if log.ErrorIfErr(err, "creating a warnings embed for user '%s' page '1", userID) {
			discutils.IError(s, i.Interaction, "Couldn't fetch the warnings :(")
			return
		}

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: emb,
			Comps: buttonFactory(userID, 1, false),
		}, discutils.I_EPHEMERAL)
	})

	discord.RegisterComponent("warnings", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		info, newPage := utils.ParsePagination(d.CustomID, super_num)
		userID := info[0]

		emb, err := warningsEmbed(userID, newPage)

		if log.ErrorIfErr(err, "creating warnings embed for user (btn) '%s' page '%d'", userID, newPage) {
			discutils.IError(s, i.Interaction, "Couldn't fetch warnings :(", discutils.I_UPDATE)
			return
		}

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: emb,
			Comps: buttonFactory(userID, newPage, false),
		}, discutils.I_UPDATE)
	})
}
