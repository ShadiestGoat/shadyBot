package donation

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
)

func autocompleteFunds(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	q := data["fund"].StringValue()
	isComplete := false
	funds, err := c.Funds("", "", false, &isComplete)
	if log.ErrorIfErr(err, "query for funds") {
		return
	}

	resp := []*discordgo.ApplicationCommandOptionChoice{}
	for _, f := range funds {
		name := f.ShortTitle + " | " + f.Title
		qWords := strings.Split(q, " ")
		count := 0.0
		for _, w := range qWords {
			if strings.Contains(name, w) {
				count++
			}
		}
		if q == "" || count > 0.65*float64(len(qWords)) {
			if len(name) > 100 {
				name = name[:97] + "..."
			}
			resp = append(resp, &discordgo.ApplicationCommandOptionChoice{
				Name:  name,
				Value: f.ID,
			})
		}
	}

	if len(resp) > 25 {
		resp = resp[:25]
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: resp,
		},
	})
}
