package misc

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func cmdEmbed() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "embed",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Description:              "Create an embed from options",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel to send this in",
				Required:    true,
				ChannelTypes: []discordgo.ChannelType{
					discordgo.ChannelTypeGuildText,
					discordgo.ChannelTypeGuildNews,
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "The embed title",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "The embed description. Use \\n for line breaks. \\\\n still counts as a line break.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "content",
				Description: "The content of a message. Use for pings. \\n for line breaks",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "desc-a",
				Description: "The embed description, in an attachment form (please only upload text!)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "img",
				Description: "The embed image, in an attachment form",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "color",
				Description: "The embed color (hex code, eg. ff00ff)",
				Required:    false,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		emb := discutils.BaseEmbed

		if color, ok := data["color"]; ok {
			if c := color.StringValue(); c != "" {
				co, err := utils.ParseColor(c)
				if err == nil {
					emb.Color = co
				}
			}
		}

		chanID := data["channel"].ChannelValue(nil).ID

		if data["title"] != nil {
			emb.Title = data["title"].StringValue()
		}

		if data["img"] != nil {
			attachmentID := data["img"].Value.(string)
			attach := d.Resolved.Attachments[attachmentID]
			emb.Image = &discordgo.MessageEmbedImage{
				URL:    attach.URL,
				Width:  attach.Width,
				Height: attach.Height,
			}
		}

		if data["desc-a"] != nil {
			attachmentID := data["desc-a"].Value.(string)
			attachment := d.Resolved.Attachments[attachmentID]
			resp, err := http.Get(attachment.URL)
			if err != nil || resp.StatusCode != 200 {
				status := "???"
				if resp != nil {
					status = fmt.Sprint(resp.StatusCode)
				}
				log.Error("Error when fetching attachment for embed desc: (err => '%v') (status: %s)", err, status)
				discutils.IError(s, i.Interaction, "Couldn't parse the description!")
				return
			}

			b, _ := io.ReadAll(resp.Body)

			if len(b) > 3950 {
				discutils.IError(s, i.Interaction, "the description was too big!")
				return
			}

			emb.Description = string(b)
		} else if data["desc"] != nil {
			desc := data["desc"].StringValue()
			desc = strings.ReplaceAll(desc, "\\n", "\n")

			emb.Description = desc
		}

		contentVal := "Preview - is this ok?"

		if data["content"] != nil {
			content := data["content"].StringValue()
			content = strings.ReplaceAll(content, "\\n", "\n")

			if len(content) > 1950 {
				discutils.IError(s, i.Interaction, "Your content is too long~")
				return
			}

			contentVal += "\n\n" + content
		}

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: &emb,
			Comps: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Confirm, send it!",
					Style:    discordgo.SuccessButton,
					CustomID: "emb_" + chanID,
				},
				discordgo.Button{
					Label:    "Lemme try again",
					Style:    discordgo.DangerButton,
					CustomID: "cancel",
				},
			},
			Content: &contentVal,
		}, discutils.I_EPHEMERAL)
	})

	discord.RegisterComponent("emb", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		spl := strings.Split(d.CustomID, "_")

		splC := strings.SplitN(i.Message.Content, "\n", 2)
		content := ""

		if len(splC) == 2 {
			content = splC[1]
		}

		msg, err := discutils.SendMessage(s, spl[1], &discordgo.MessageSend{
			Content: content,
			Embeds:  i.Message.Embeds,
		})

		if log.ErrorIfErr(err, "sending embed after btn") {
			discutils.IError(s, i.Interaction, "Sorry, couldn't send your embed :(", discutils.I_UPDATE)
			return
		}

		emb := discutils.BaseEmbed
		emb.Title = "Success!"
		emb.Description = "Embed sent [Here](" + discutils.EMessageURL(msg) + "] <3"

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_UPDATE|discutils.I_EPHEMERAL)
	})
}
