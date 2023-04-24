package polls

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/characters"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func modal() {
	discord.RegisterModal("poll", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ModalSubmitInteractionData, data map[string]string) {
		split := strings.SplitN(d.CustomID, "_", 2)
		otherOpt := split[1]

		optFields := ""

		optMax := 0

		for i, letter := range utils.Alphabet {
			optionContent := data[letter]
			if optionContent == "" {
				optMax = i
				break
			}

			optFields += fmt.Sprintf(":regional_indicator_%s: %s\n", letter, optionContent)
		}

		if otherOpt != "" {
			optFields += ":o2: " + otherOpt + "\n"
		}

		optFields = optFields[:len(optFields)-1]

		emb := discutils.BaseEmbed
		emb.Title = "📊 Poll - " + data["title"]
		emb.Color = discutils.COLOR_PRIMARY
		emb.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "Options",
				Value: optFields,
			},
		}

		msg, err := discutils.SendMessage(s, config.Channels.Polls, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				&emb,
			},
		})

		if log.ErrorIfErr(err, "sending poll msg") {
			discutils.IError(s, i.Interaction, "Couldn't send poll message :(")
			return
		}

		respEmb := discutils.BaseEmbed
		respEmb.Color = discutils.COLOR_PRIMARY
		respEmb.Title = "Success! The poll was created!"
		respEmb.Description = fmt.Sprintf("[Check it out!](%v)", discutils.EMessageURL(msg))

		discutils.IEmbed(s, i.Interaction, &respEmb, discutils.I_NONE)

		for i := 0; i < optMax; i++ {
			s.MessageReactionAdd(config.Channels.Polls, msg.ID, characters.AlphabetEmotes[i])
		}

		if otherOpt != "" {
			s.MessageThreadStartComplex(config.Channels.Polls, msg.ID, &discordgo.ThreadStart{
				Name:                otherOpt,
				AutoArchiveDuration: 24 * 60,
				Type:                discordgo.ChannelTypeGuildText,
			})

			s.MessageReactionAdd(config.Channels.Polls, msg.ID, "🅾️")
		}
	})
}
