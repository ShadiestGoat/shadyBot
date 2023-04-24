package warnings

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func cmdPardon() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "pardon",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Description:              "Pardon a user's warning",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "id",
				Description: "The ID of the warning to pardon (use /warnings to get this)",
				Required:    true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		id := data["id"].StringValue()

		warning := Warning{
			ID:         id,
			WarnedUser: "",
			Severity:   0,
			Reason:     "",
		}

		err := db.QueryRowID("SELECT warned_user, severity, reason FROM warnings WHERE id = $1", id, &warning.WarnedUser, &warning.Severity, &warning.Reason)

		if err != nil {
			if db.NoRows(err) {
				discutils.IError(s, i.Interaction, "This warning does not exist! Use "+discord.CommandMention("warnings", "")+" to find the warnings that you need")
			} else {
				discutils.IError(s, i.Interaction, "DB Error :(")
			}
			return
		}

		emb := discutils.BaseEmbed
		emb.Title = "Warning Pardoned"
		emb.Description = "<@" + warning.WarnedUser + ">'s warning for `" + strings.ReplaceAll(warning.Reason, "`", "\\`") + "` has been pardoned. Their warning severity dropped by " + fmt.Sprint(warning.Severity) + "!"

		_, err = db.Exec(`DELETE FROM warnings WHERE id = $1`, id)

		if err != nil {
			discutils.IError(s, i.Interaction, "Error when deleting the warning ://")
			return
		}

		newTotSev, newTmpSev := Severity(warning.WarnedUser)

		oldPun := Punishment(newTmpSev+warning.Severity, newTotSev+warning.Severity)
		newPun := Punishment(newTmpSev, newTotSev)

		if !newPun.Ban && oldPun.Ban {
			s.GuildBanDelete(config.Discord.GuildID, warning.WarnedUser)
		}

		mem, err := s.GuildMember(config.Discord.GuildID, warning.WarnedUser)

		if err == nil && mem != nil {
			if mem.CommunicationDisabledUntil != nil && time.Until(*mem.CommunicationDisabledUntil) > 0 {
				left := time.Until(*mem.CommunicationDisabledUntil)
				newTimeout := newPun.Timeout - (oldPun.Timeout - left)
				timeoutUntil := time.Now().Add(newTimeout)
				if newTimeout > 0 {
					s.GuildMemberEditComplex(config.Discord.GuildID, warning.WarnedUser, &discordgo.GuildMemberParams{
						CommunicationDisabledUntil: &timeoutUntil,
					})
				}
			}
		}

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_EPHEMERAL)

		emb2 := discutils.BaseEmbed
		emb2.Title = "⚠️ A warning given to you has been lifted!"

		discutils.DM(s, warning.WarnedUser, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				&emb2,
			},
		})

		if config.Channels.ModLog != "" {
			discutils.SendMessage(s, config.Channels.ModLog, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					&emb,
				},
			})
		}
	})
}
