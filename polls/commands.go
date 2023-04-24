package polls

import (
	"fmt"
	"math"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/characters"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func cmd() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.ChatApplicationCommand,
		Name:              "poll",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Description:       "Create a poll",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "other",
				Description: "The 'other' option's name",
				MaxLength:   18,
				Required:    false,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		customID := "poll_"
		if opt, ok := data["other"]; ok {
			customID += opt.StringValue()
		}

		comps := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "title",
						Label:       "Give a title to this poll",
						Style:       discordgo.TextInputShort,
						Placeholder: "Title",
						Required:    true,
						MinLength:   0,
						MaxLength:   60,
					},
				},
			},
		}

		for i := 0; i < 4; i++ {
			comps = append(comps, discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:  utils.Alphabet[i],
						Label:     "Option #" + fmt.Sprint(i+1),
						Style:     discordgo.TextInputShort,
						Required:  i < 2,
						MaxLength: 48,
					},
				},
			})
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				Title:      "New Poll",
				CustomID:   customID,
				Components: comps,
			},
		})
	})

	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.MessageApplicationCommand,
		Name:              "End Poll",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		ogMsg := d.Resolved.Messages[d.TargetID]

		if ogMsg.ChannelID != config.Channels.Polls {
			discutils.IError(s, i.Interaction, "Only do this on actual polls!")
			return
		}

		if ogMsg.Embeds[0].Description != "" {
			discutils.IError(s, i.Interaction, "This poll has already closed!")
			return
		}

		newEmb := ogMsg.Embeds[0]
		newEmb.Description = "**This poll has been concluded!**\n"
		newEmb.Color = discutils.COLOR_SUCCESS

		spl := strings.Split(newEmb.Fields[0].Value, "\n")

		newOpts := ""

		votes := make([]float64, len(spl))
		total := 0.0

		best := []int{
			0,
		}

		for i, opt := range spl {
			emote := characters.AlphabetEmotes[i]

			if opt[1] == 'o' {
				emote = "🅾️"
			}

			count := 0.0

			for _, r := range ogMsg.Reactions {
				if r.Emoji.Name == emote {
					count = float64(r.Count) - 1
					break
				}
			}

			if count > votes[best[0]] {
				best = []int{i}
			}
			if count == votes[best[0]] {
				best = append(best, i)
			}

			votes[i] = count
			total += count
		}

		for i, opt := range spl {
			effect := ""
			for _, b := range best {
				if b == i {
					effect = "**"
					break
				}
			}
			str := opt
			if opt[len(opt)-1] == '\n' {
				str = str[:len(opt)-1]
			}
			voteDisplay := 1000 * (votes[i] / total)
			voteDisplay = math.Round(voteDisplay) / 10
			newOpts += effect + str + fmt.Sprintf(" - %.1f%%%v\n", voteDisplay, effect)
		}

		newEmb.Fields[0].Value = newOpts

		s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: config.Channels.Polls,
			ID:      d.TargetID,

			Embeds: []*discordgo.MessageEmbed{
				newEmb,
			},
		})

		emb := discutils.BaseEmbed
		emb.Title = "Poll Ended Successfully (you're such a good girl <3)"

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_EPHEMERAL)

		archive := true
		s.ChannelEditComplex(d.TargetID, &discordgo.ChannelEdit{
			Archived: &archive,
			Locked:   &archive,
		})
	})
}
