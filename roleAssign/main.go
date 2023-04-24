package roleassign

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func register() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.ChatApplicationCommand,
		Name:              "role",
		Description:       "Create a role assignment button",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "The role to assign",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "The description of the assignment embed",
				Required:    true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		role := data["role"].RoleValue(s, config.Discord.GuildID)

		emb := discutils.BaseEmbed
		emb.Title = "The '" + role.Name + "' Role"
		emb.Description = data["desc"].StringValue()

		msg, err := discutils.SendMessage(s, config.Channels.RoleAssignment, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				&emb,
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Get/Remove this role",
							Style:    discordgo.PrimaryButton,
							CustomID: "role_" + role.ID,
						},
					},
				},
			},
		})

		if log.ErrorIfErr(err, "sending role message") {
			discutils.IError(s, i.Interaction, "Couldn't send the message (sorry)")
			return
		}

		embNew := discutils.BaseEmbed

		embNew.Title = "The role assignment was created!"
		embNew.Description = fmt.Sprintf("[Take a look](%s)", discutils.EMessageURL(msg))

		discutils.IEmbed(s, i.Interaction, &embNew, discutils.I_EPHEMERAL)
	})

	discord.RegisterComponent("role", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		if i.Member == nil {
			discutils.IError(s, i.Interaction, "This can only be executed inside a guild!")
			return
		}

		roleID := strings.Split(d.CustomID, "_")[1]

		newRoles := []string{}

		for j, r := range i.Member.Roles {
			if r == roleID {
				newRoles = append(i.Member.Roles[:j], i.Member.Roles[j+1:]...)
				break
			}
		}

		shouldAddRole := len(newRoles) == 0

		if shouldAddRole {
			newRoles = append(i.Member.Roles, roleID)
		}

		_, err := s.GuildMemberEditComplex(config.Discord.GuildID, i.Member.User.ID, &discordgo.GuildMemberParams{
			Roles: &newRoles,
		})

		if log.ErrorIfErr(err, "updating member roles for role btn") {
			discutils.IError(s, i.Interaction, "Couldn't set your roles, sorry ://")
			return
		}

		emb := discutils.BaseEmbed

		if shouldAddRole {
			emb.Title = "The role was added!"
			emb.Description = "You now have the <@&" + roleID + "> role ^^"
		} else {
			emb.Title = "The role was removed!"
			emb.Description = "You not longer have the <@&" + roleID + "> role ^^"
		}

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_EPHEMERAL)
	})
}
