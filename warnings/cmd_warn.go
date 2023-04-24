package warnings

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/snownode"
)

func cmdWarn() {
	minSeverity := 1.0

	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.ChatApplicationCommand,
		Name:              "warn",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Description:       "Warn a user",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to warn",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "The reason for the warning",
				Required:    true,
				MaxLength:   128,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "severity",
				Description: "Severity of the warning (default = 1)",
				Required:    false,
				MinValue:    &minSeverity,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		warningID := snownode.Generate()
		severity := 1
		if data["severity"] != nil {
			severity = int(data["severity"].IntValue())
		}
		reason := data["reason"].StringValue()
		userID := data["user"].UserValue(nil).ID

		_, err := db.Exec(`INSERT INTO warnings(id, warned_user, severity, reason) VALUES ($1, $2, $3, $4)`, warningID, userID, severity, reason)

		if log.ErrorIfErr(err, "inserting warning") {
			discutils.IError(s, i.Interaction, "Couldn't insert the warning!")
			return
		}

		emb := discutils.BaseEmbed
		emb.Title = "⚠️ You were warned ⚠️"

		severityTot, severityTmp := Severity(userID)

		punishment := Punishment(severityTmp, severityTot)

		emb.Description = punishment.Msg

		if punishment.Timeout != 0 {
			punishUntil := time.Now().Add(punishment.Timeout)
			s.GuildMemberEditComplex(config.Discord.GuildID, userID, &discordgo.GuildMemberParams{
				CommunicationDisabledUntil: &punishUntil,
			})
		}

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Punishment",
			Value:  punishment.String(),
			Inline: false,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Reason",
			Value:  reason,
			Inline: false,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Total Warnings (severity based)",
			Value:  fmt.Sprint(severityTot),
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  discutils.EMBED_EMPTY,
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Temporary Warnings (severity based)",
			Value:  fmt.Sprint(severityTmp),
			Inline: true,
		})

		amountTot, amountTmp := Amount(userID)

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Total Warning Amount",
			Value:  fmt.Sprint(amountTot),
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   discutils.EMBED_EMPTY,
			Value:  discutils.EMBED_EMPTY,
			Inline: true,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Temporary Warning Amount",
			Value:  fmt.Sprint(amountTmp),
			Inline: true,
		})

		discutils.DM(s, userID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				&emb,
			},
		})

		if punishment.Ban {
			s.GuildBan(config.Discord.GuildID, userID)
		}

		if config.Channels.ModLog != "" {
			emb3 := emb

			newFields := append([]*discordgo.MessageEmbedField{}, emb3.Fields...)

			emb3.Fields = newFields

			emb3.Fields = append(emb3.Fields, &discordgo.MessageEmbedField{
				Name:   "User",
				Value:  "<@" + userID + ">",
				Inline: true,
			})

			emb3.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   discutils.EMBED_EMPTY,
				Value:  discutils.EMBED_EMPTY,
				Inline: true,
			})

			emb3.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   "ID",
				Value:  warningID,
				Inline: true,
			})

			discutils.SendMessage(s, config.Channels.ModLog, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					&emb3,
				},
			})
		}

		emb2 := discutils.BaseEmbed

		emb2.Title = "Success"
		emb2.Description = "You have successfully warned <@" + userID + ">:\n\n```\n" + reason + "\n```\nPunishment: " + punishment.String()

		discutils.IEmbed(s, i.Interaction, &emb2, discutils.I_EPHEMERAL)
	})
}
