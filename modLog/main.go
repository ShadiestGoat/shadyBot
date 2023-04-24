package modlog

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/initializer"
)

func init() {
	initializer.Register(initializer.MOD_MOD_LOG, func(c *initializer.InitContext) {
		discord.MessageUpdate.Add(func(s *discordgo.Session, e *discordgo.MessageUpdate) bool {
			if e.GuildID == "" || e.Author == nil || e.Author.Bot || e.Message == nil {
				return true
			}
			return false
		})

		discord.MessageUpdate.Add(func(s *discordgo.Session, e *discordgo.MessageUpdate) bool {
			embeds := []*discordgo.MessageEmbed{}

			if e.BeforeUpdate == nil {
				emb := discutils.BaseEmbed
				emb.Title = "Message Edit"
				emb.Description = "We do not know what it was like before, but, the new message content is this:\n```md\n" + e.Content + "\n```"
				embeds = append(embeds, &emb)
			} else {
				emb := discutils.BaseEmbed
				emb.Title = "Message Edit (1/3)"
				emb.Fields = []*discordgo.MessageEmbedField{
					{
						Name:   "Author",
						Value:  "<@" + e.Author.ID + ">",
						Inline: true,
					},
					{
						Name:   "Channel",
						Value:  "<#" + e.Message.ChannelID + ">",
						Inline: true,
					},
					{
						Name:   "Message URL",
						Value:  "[Here](" + discutils.EMessageURL(e.Message) + ")",
						Inline: true,
					},
				}
				embeds = append(embeds, &emb)

				emb2 := discutils.BaseEmbed
				emb2.Title = "Message Edit (2/3)"
				emb2.Description = "Old Content:\n```md\n" + e.BeforeUpdate.Content + "\n```"

				oldAttachments := discutils.AttachmentsString(e.BeforeUpdate)
				if oldAttachments != "" {
					emb2.Fields = append(emb2.Fields, &discordgo.MessageEmbedField{
						Name:   "Attachments",
						Value:  oldAttachments,
						Inline: false,
					})
				}

				embeds = append(embeds, &emb2)

				emb3 := discutils.BaseEmbed
				emb3.Title = "Message Edit (3/3)"
				emb3.Description = "New Content:\n```md\n" + e.Content + "\n```"

				newAttachments := discutils.AttachmentsString(e.Message)
				if newAttachments != "" {
					emb3.Fields = append(emb3.Fields, &discordgo.MessageEmbedField{
						Name:   "Attachments",
						Value:  newAttachments,
						Inline: false,
					})
				}

				embeds = append(embeds, &emb3)
			}

			discutils.SendMessage(s, config.Channels.ModLog, &discordgo.MessageSend{
				Embeds: embeds,
			})

			return false
		})

		discord.MessageRemove.Add(func(s *discordgo.Session, e *discordgo.MessageDelete) bool {
			if e.BeforeDelete == nil {
				return false
			}
			if e.BeforeDelete.Author == nil || e.BeforeDelete.Author.Bot {
				return false
			}
			if e.BeforeDelete.GuildID == "" {
				return false
			}

			emb := discutils.BaseEmbed
			emb.Title = "Message Delete"
			emb.Fields = []*discordgo.MessageEmbedField{
				{
					Name:   "Message Author",
					Value:  "<@" + e.Author.ID + ">",
					Inline: true,
				},
				{
					Name:   "Channel",
					Value:  "<#" + e.BeforeDelete.ChannelID + ">",
					Inline: true,
				},
				{
					Name:   "Message ID",
					Value:  e.BeforeDelete.ID,
					Inline: true,
				},
			}

			attachments := discutils.AttachmentsString(e.BeforeDelete)

			if attachments != "" {
				emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
					Name:   "Attachments",
					Value:  attachments,
					Inline: false,
				})
			}

			emb.Description = e.BeforeDelete.Content

			discutils.SendMessage(s, config.Channels.ModLog, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					&emb,
				},
			})

			return false
		})
	}, &initializer.ModuleInfo{
		ConfigOpts: []*string{
			&config.Channels.ModLog,
		},
		PreHooks: []initutils.Module{initializer.MOD_HELP_LOADER},
	})
}
