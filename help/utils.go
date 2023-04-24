package help

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/characters"
	"github.com/shadiestgoat/shadyBot/discutils"
)

var sections = []*sectionStage2{}
var sectionMap = map[string][]string{}

func helpResp(s *discordgo.Session, i *discordgo.Interaction, secName string, isBtn bool) {
	docs := sectionMap[secName]

	embeds := []*discordgo.MessageEmbed{}

	for _, d := range docs {
		emb := discutils.BaseEmbed
		emb.Title = secName
		emb.Description = d

		embeds = append(embeds, &emb)
	}

	curI := 0

	for j, s := range sections {
		if s.Name == secName {
			curI = j
			break
		}
	}

	opts := discutils.I_EPHEMERAL

	if isBtn {
		opts |= discutils.I_UPDATE
	}

	discutils.IResp(s, i, &discutils.IRespOpts{
		Embeds: embeds,
		Comps: []discordgo.MessageComponent{
			discordgo.Button{
				Style: discordgo.PrimaryButton,
				Emoji: discordgo.ComponentEmoji{
					Name: characters.ARROW_LL,
				},
				Disabled: curI == 0,
				CustomID: "help_prev_" + fmt.Sprint(curI),
			},
			discordgo.Button{
				Style: discordgo.PrimaryButton,
				Emoji: discordgo.ComponentEmoji{
					Name: characters.ARROW_RR,
				},
				Disabled: curI == len(sections)-1,
				CustomID: "help_next_" + fmt.Sprint(curI),
			},
		},
	}, opts)
}
