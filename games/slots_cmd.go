package games

import (
	"fmt"
	"math"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/characters"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
	"github.com/shadiestgoat/shadyBot/xp"
)

var slotHorizontalLineReq = [3]int{2, 1, 3}

const LINE_COST = 50

func cmdSlots() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "slots",
		Description: "Spend XP on spinning the slots",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "paylines",
				Description: "The number of paylines you are buying (Default = 1)",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Center Only (1)",
						Value: 1,
					},
					{
						Name:  "Center & Top Lines (2)",
						Value: 2,
					},
					{
						Name:  "Horizontal Lines (3)",
						Value: 3,
					},
					{
						Name:  "All Horizontal and 1 Diagonal Lines (4)",
						Value: 4,
					},
					{
						Name:  "All Lines (5)",
						Value: 5,
					},
				},
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		lines := 1

		if data["paylines"] != nil {
			lines = int(data["paylines"].IntValue())
		}

		if InActivityErrorCheck(s, i.Interaction) {
			return
		}
		if XPErrorCheck(s, i.Interaction, lines*LINE_COST) {
			return
		}
		ActivityStore.Add(i.Interaction, GT_SLOTS)

		cols := [3][]rune{slots(), slots(), slots()}
		shiftsLeft := [3]int{
			utils.RandInt(5, 17),
			utils.RandInt(5, 17),
			utils.RandInt(5, 17),
		}

		maxV := 0
		for _, v := range shiftsLeft {
			if v > maxV {
				maxV = v
			}
		}

		userID := discutils.IAuthor(i.Interaction).ID

		content := "<@" + userID + ">"

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed:   embedSlots(lines, cols, shiftsLeft),
			Content: &content,
		}, discutils.I_NONE)

		time.Sleep(120 * time.Millisecond)

		for j := 0; j < maxV; j++ {
			for k := 0; k < 3; k++ {
				if shiftsLeft[0] != 0 {
					shiftsLeft[0]--
					cols[k] = utils.RingCircleMove(cols[k])
				}
			}

			discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
				Embed:   embedSlots(lines, cols, shiftsLeft),
				Content: &content,
			}, discutils.I_EDIT)

			time.Sleep(130 * time.Millisecond)
		}

		emb, total := embedSlotsWithRewards(lines, cols)

		if total >= SLOTS_JACKPOT {
			xp.XPEventChan <- &xp.XPEventInfo{
				Event:   xp.EV_JACKPOT,
				IntInfo: SLOTS_JACKPOT,
				UserID:  userID,
			}
		}

		FinishGame(userID, int(math.Abs(float64(total)-float64(lines)*LINE_COST)), total > 0, GT_SLOTS)

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed:   emb,
			Content: &content,
		}, discutils.I_EDIT)
	})
}

// Use *char* if *cond*, otherwise use GAME_EMPTY_CHAR
func character(cond bool, char string) string {
	if cond {
		return char
	}

	return string(characters.GAME_EMPTY_CHAR)
}

const slotPayouts = "```\n" + `🟪 = 🟦 = 🟫 = 🟩 = 🟥 = 🟨 = 🟧

ANY 😳   = 54
ANY ❗   = 69
🟪 🟪    = 34
😳 😳    = 505
❗ ❗    = 1984
🟪 🟪 🟪 = 419
😳 😳 😳 = 11,111
❗ ❗ ❗ = 66,666
` + "```"

func embedSlots(lines int, cols [3][]rune, shiftsLeft [3]int) *discordgo.MessageEmbed {
	square := character(lines >= 4, characters.ARROW_DR) +
		character(shiftsLeft[0] != 0, string(characters.SQR_BORDER_BLACK)) +
		character(shiftsLeft[1] != 0, string(characters.SQR_BORDER_BLACK)) +
		character(shiftsLeft[2] != 0, string(characters.SQR_BORDER_BLACK)) +
		character(lines >= 5, characters.ARROW_DL) +
		"\n"

	for i := 0; i < 3; i++ {
		displayArrow := lines >= slotHorizontalLineReq[i]

		square += character(displayArrow, characters.ARROW_RR) +
			string(cols[0][i]) + string(cols[1][i]) + string(cols[2][i]) +
			character(displayArrow, characters.ARROW_LL) +
			"\n"
	}

	square += character(lines >= 4, characters.ARROW_UR) +
		character(shiftsLeft[0] != 0, string(characters.SQR_BORDER_BLACK)) +
		character(shiftsLeft[1] != 0, string(characters.SQR_BORDER_BLACK)) +
		character(shiftsLeft[2] != 0, string(characters.SQR_BORDER_BLACK)) +
		character(lines >= 5, characters.ARROW_UL) +
		"\n"

	emb := discutils.BaseEmbed
	emb.Title = TITLE_SLOTS
	emb.Description = square
	emb.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "Payouts",
			Value:  slotPayouts,
			Inline: false,
		},
	}

	return &emb
}

func embedSlotsWithRewards(lines int, cols [3][]rune) (*discordgo.MessageEmbed, int) {
	emb := embedSlots(lines, cols, [3]int{})

	rewards := [5]int{-1, -1, -1, -1, -1}

	for i := 0; i < lines; i++ {
		var rawCols [3]rune

		switch i {
		case 0:
			rawCols = [3]rune{
				cols[0][1],
				cols[1][1],
				cols[2][1],
			}
		case 1:
			rawCols = [3]rune{
				cols[0][0],
				cols[1][0],
				cols[2][0],
			}
		case 2:
			rawCols = [3]rune{
				cols[0][2],
				cols[1][2],
				cols[2][2],
			}
		case 3:
			rawCols = [3]rune{
				cols[0][0],
				cols[1][1],
				cols[2][2],
			}
		case 4:
			rawCols = [3]rune{
				cols[0][2],
				cols[1][1],
				cols[2][0],
			}
		}
		
		rewards[i] = slotsRewards(rawCols)
	}

	rewardStr := ""
	total := 0

	for j, r := range rewards {
		rewardStr += fmt.Sprintf("Line %d: ", j+1)
		if r == -1 {
			rewardStr += "🔒"
		} else if r == 0 {
			rewardStr += "No reward :("
		} else {
			total += r
			rewardStr += fmt.Sprint(r)
		}
		rewardStr += "\n"
	}

	rewardStr += "████████████\n**Total: " + fmt.Sprint(total-lines*LINE_COST) + "**"

	emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
		Name:   "Rewards",
		Value:  rewardStr,
		Inline: false,
	})

	return emb, total
}
