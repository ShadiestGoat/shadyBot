package purge

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func init() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.ChatApplicationCommand,
		Name:              "purge",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Description:       "Fully purge a channel (WARNING: THIS RE-CREATES THE CHANNEL)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel to purge",
				Required:    true,
				ChannelTypes: []discordgo.ChannelType{
					discordgo.ChannelTypeGuildText,
				},
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		chanID := data["chan"].ChannelValue(nil).ID

		emb := discutils.BaseEmbed

		emb.Title = "Channel Purge"
		emb.Description = "This will delete all messages in <#" + chanID + ">.\nThis will also re-create the channel. Make sure any configurations are changed.\n**Are you sure you want to do this?**"

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: &emb,
			Comps: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Yes, purge it!",
					Style:    discordgo.PrimaryButton,
					CustomID: "purge_" + chanID,
				},

				discordgo.Button{
					Label:    "Cancel!",
					Style:    discordgo.DangerButton,
					CustomID: "cancel",
				},
			},
		}, discutils.I_EPHEMERAL)
	})

	discord.RegisterComponent("purge", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		chanID := strings.Split(d.CustomID, "_")[1]

		oldChan, err := s.Channel(chanID)

		if log.ErrorIfErr(err, "fetching channel for purge") {
			discutils.IError(s, i.Interaction, "Couldn't fetch channel", discutils.I_UPDATE)
			return
		}

		newChan, err := s.GuildChannelCreateComplex(config.Discord.GuildID, discordgo.GuildChannelCreateData{
			Name:                 oldChan.Name,
			Type:                 oldChan.Type,
			Topic:                oldChan.Topic,
			Bitrate:              oldChan.Bitrate,
			UserLimit:            oldChan.UserLimit,
			RateLimitPerUser:     oldChan.RateLimitPerUser,
			Position:             oldChan.Position,
			PermissionOverwrites: oldChan.PermissionOverwrites,
			ParentID:             oldChan.ParentID,
			NSFW:                 oldChan.NSFW,
		})

		if log.ErrorIfErr(err, "fetching channel for purge") {
			discutils.IError(s, i.Interaction, "Couldn't create the new channel", discutils.I_UPDATE)
			return
		}

		_, err = s.ChannelDelete(chanID)
		if log.ErrorIfErr(err, "deleting channel for purge") {
			discutils.IError(s, i.Interaction, "Couldn't delete the old channel", discutils.I_UPDATE)
			return
		}

		emb := discutils.BaseEmbed
		emb.Title = "Channel purge complete"
		emb.Description = "The channel now resides in <#" + newChan.ID + ">!"

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: &emb,
			Comps: []discordgo.MessageComponent{},
		}, discutils.I_UPDATE|discutils.I_EPHEMERAL)
	})
}
