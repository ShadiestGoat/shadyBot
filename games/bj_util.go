package games

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

// An array of BJ cards, which is 2-14, where ace is 14
type BJHand []int

type BJGame struct {
	UserID     string
	Bet        int
	Doubled    bool
	UserHand   BJHand
	DealerHand BJHand
	UserTurn   bool
}

func (g *BJGame) NewCard() int {
	curCards := []int{}
	curCards = append(curCards, g.UserHand...)
	curCards = append(curCards, g.DealerHand...)

	badM := map[int]int{}
	for _, c := range curCards {
		badM[c]++
	}

	availableCards := []int{}

	for j := 0; j < 4; j++ {
		for i := 2; i <= 14; i++ {
			if badM[i] > j {
				continue
			}
			availableCards = append(availableCards, i)
		}
	}

	return availableCards[utils.RandInt(0, len(availableCards)-1)]
}

func (h BJHand) Totals(filterBust bool) []int {
	totals := []int{0}

	for _, v := range h {
		newTots := []int{}
		if v > 10 && v != 14 {
			v = 10
		}

		for i, t := range totals {
			if v == 14 {
				newTots = append(newTots, t+1)
				totals[i] += 11
			} else {
				totals[i] += v
			}
		}

		totals = append(totals, newTots...)
	}

	totM := map[int]bool{}

	newTots := []int{}

	for _, t := range totals {
		if !totM[t] && (t <= 21 || !filterBust) {
			newTots = append(newTots, t)
		}

		totM[t] = true
	}

	sort.Slice(newTots, func(i, j int) bool {
		return newTots[i] > newTots[j]
	})

	return newTots
}

func (g BJGame) TrueBet() int {
	if g.Doubled {
		return g.Bet * 2
	}

	return g.Bet
}

// Parses a BJ game from an embed. Can return nil! That indicates an error in parsing.
func bjParseGame(emb *discordgo.MessageEmbed) *BJGame {
	if len(emb.Fields) != 3 {
		return nil
	}

	game := &BJGame{}

	betField := emb.Fields[0].Value

	rawBet := []rune{}
	for _, b := range betField {
		if b < '0' || b > '9' {
			break
		}
		rawBet = append(rawBet, b)
	}

	bet, err := strconv.Atoi(string(rawBet))
	if err != nil {
		log.Warn("Couldn't parse BJ game, parsing bet '%s'", string(rawBet))
		return nil
	}
	game.Bet = bet

	game.Doubled = betField[len(betField)-1] == ')'

	lines := strings.Split(emb.Description, "\n")

	userLine := lines[0]

	for _, b := range userLine {
		if b >= '0' && b <= '9' {
			game.UserID += string(b)
		}
	}
	game.UserTurn = lines[len(lines)-1][6] == 'U'

	game.UserHand = bjParseHand(emb.Fields[2].Value)
	game.DealerHand = bjParseHand(emb.Fields[1].Value)

	return game
}

// parses a BJ Hand from a string
func bjParseHand(inp string) BJHand {
	inp = strings.SplitN(inp, " |", 2)[0]
	spl := strings.Split(inp, ", ")

	hand := BJHand{}

	for _, card := range spl {
		v := 0
		switch card[0] {
		case 'J':
			v = 11
		case 'Q':
			v = 12
		case 'K':
			v = 13
		case 'A':
			v = 14
		default:
			p, _ := strconv.Atoi(card)
			v = p
		}
		hand = append(hand, v)
	}

	return hand
}

type bjDealerState int

const (
	bjd_bust bjDealerState = iota
	bjd_draw
	bjd_continue
	bjd_lost
)

func (g BJGame) dealerLoopState() bjDealerState {
	dealerTots := g.DealerHand.Totals(true)
	userTots := g.UserHand.Totals(true)

	if len(dealerTots) == 0 {
		return bjd_bust
	}

	if dealerTots[0] == userTots[0] && dealerTots[0] >= 17 {
		return bjd_draw
	}

	if dealerTots[0] <= userTots[0] && dealerTots[len(dealerTots)-1] < 17 {
		return bjd_continue
	}

	return bjd_lost
}

//
// EMBED FORMATS
//

func (g BJGame) embedBase(desc string, won bool) *discordgo.MessageEmbed {
	emb := discutils.BaseEmbed

	emb.Title = TITLE_BJ

	desc += "! "

	if g.Doubled {
		desc += "Your original bet was " + fmt.Sprint(g.Bet) + ", but you doubled down so now it is **" + fmt.Sprint(g.Bet*2) + "**! You **"
	} else {
		desc += "Your bet was **" + fmt.Sprint(g.Bet) + "**, and you **"
	}

	if won {
		desc += "won"
		emb.Color = discutils.COLOR_SUCCESS
	} else {
		desc += "lost"
		emb.Color = discutils.COLOR_DANGER
	}

	desc += "** this xp." + g.handString()

	emb.Description = desc

	return &emb
}

func (g BJGame) embedGame() *discordgo.MessageEmbed {
	turn := "Dealer's"

	if g.UserTurn {
		turn = "User's"
	}

	betSuffix := ""

	if g.Doubled {
		betSuffix = " (x2)"
	}

	emb := discutils.BaseEmbed
	emb.Title = TITLE_BJ
	emb.Description = "User: <@" + g.UserID + ">\n\nTurn: " + turn

	dealerSuffix := ""

	if !g.UserTurn {
		dealerSuffix = " <a:loading:1019317568651145246>"
	}

	emb.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "Bet",
			Value:  fmt.Sprint(g.Bet) + betSuffix,
			Inline: false,
		},
		{
			Name:   "Dealer's Hand",
			Value:  g.DealerHand.String() + dealerSuffix,
			Inline: true,
		},
		{
			Name:   "Your Hand",
			Value:  g.UserHand.String(),
			Inline: true,
		},
	}

	return &emb
}

//
// STRING FORMATS
//

func (g BJGame) handString() string {
	return "\n\n<@" + g.UserID + ">'s hand: " + g.UserHand.String() + "\nDealer's hand: " + g.DealerHand.String()
}

func (h BJHand) String() string {
	str := ""

	for _, v := range h {
		switch v {
		case 11:
			str += "J (10)"
		case 12:
			str += "Q (10)"
		case 13:
			str += "K (10)"
		case 14:
			str += "A (1/11)"
		default:
			str += fmt.Sprint(v)
		}
		str += ", "
	}

	str = str[:len(str)-2] + " | "

	tots := h.Totals(false)
	foundBig := false

	for _, v := range tots {
		effect := ""
		if v <= 21 && !foundBig {
			effect = "**"
			foundBig = true
		}

		str += effect + fmt.Sprint(v) + effect + "/"
	}

	return str[:len(str)-1]
}
