package games

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/xp"
)

const (
	TITLE_BJ    = "🎮 - Black Jack"
	TITLE_SLOTS = "🎮 - Slots"
	TITLE_COIN  = "🎮 - Coin Flip"
)

type InactivityDB struct {
	*sync.Mutex
	Info map[string]*GambleInactivityInfo
}

var ActivityStore = &InactivityDB{
	Mutex: &sync.Mutex{},
	Info:  map[string]*GambleInactivityInfo{},
}

type GameType int

const (
	GT_BJ GameType = iota
	GT_SLOTS
	GT_COIN
)

func (g GameType) String() string {
	switch g {
	case GT_BJ:
		return "blackjack"
	case GT_SLOTS:
		return "slots"
	case GT_COIN:
		return "coin flip"
	}

	return ""
}

type GambleInactivityInfo struct {
	Time            time.Time
	Game            GameType
	LastInteraction *discordgo.Interaction
}

func (g *InactivityDB) Remove(userID string) {
	g.Lock()
	defer g.Unlock()

	delete(g.Info, userID)
}

// Returns true if the user is alr in an active game
func (g *InactivityDB) UserIsInGame(i *discordgo.Interaction) bool {
	g.Lock()
	defer g.Unlock()

	_, ok := g.Info[discutils.IAuthor(i).ID]
	return ok
}

func (g *InactivityDB) Add(i *discordgo.Interaction, game GameType) {
	g.Lock()
	defer g.Unlock()

	g.Info[discutils.IAuthor(i).ID] = &GambleInactivityInfo{
		Time:            time.Now(),
		Game:            game,
		LastInteraction: i,
	}
}

func (g *InactivityDB) Update(i *discordgo.Interaction) {
	g.Lock()
	defer g.Unlock()
	if g, ok := g.Info[discutils.IAuthor(i).ID]; ok {
		g.Time = time.Now()
		g.LastInteraction = i
	}
}

func (g *InactivityDB) Loop(s *discordgo.Session, stop chan bool) {
	ticker := time.NewTicker(7 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			go func() {
				g.Lock()
				defer g.Unlock()

				for _, t := range g.Info {
					if time.Since(t.Time) >= 3*time.Minute {
						if t.Game == GT_SLOTS {
							log.Warn("Slots inactivity leak!")
							continue
						}

						msg, err := s.InteractionResponse(t.LastInteraction)
						bet := 0
						user := ""

						if msg != nil && len(msg.Embeds) == 1 && err == nil {
							switch t.Game {
							case GT_BJ:
								game := bjParseGame(msg.Embeds[0])
								if game != nil {
									bet = game.TrueBet()
									user = game.UserID
								}
							}
						}

						if user != " " {
							go FinishGame(user, bet, false, GT_BJ)
						}

						discutils.IError(s, t.LastInteraction, fmt.Sprintf("The game timed out! You lost your bet, %d!", bet), discutils.I_DEFERRED)
					}
				}
			}()
		}
	}
}

// Checks & Responds to active gambling game. Returns true if in an active game
func InActivityErrorCheck(s *discordgo.Session, i *discordgo.Interaction) bool {
	if ActivityStore.UserIsInGame(i) {
		discutils.IError(s, i, "You are already in a game!")
		return true
	}
	return false
}

// Checks & Responds to error if user doesn't have enough xp. Returns true if there is an error!
func XPErrorCheck(s *discordgo.Session, i *discordgo.Interaction, bet int) bool {
	userXP, err := xp.FetchXP(discutils.IAuthor(i).ID)
	if err != nil {
		discutils.IError(s, i, "Couldn't fetch your xp :(")
		return true
	}

	if !userXP.HasNeededXP(bet) {
		discutils.IError(s, i, "Your bet was too large! You are too broke!")
		return true
	}
	return false
}

// Use bet = 0 for draws
func FinishGame(userID string, bet int, won bool, game GameType) {
	if bet != 0 {
		trueBet := bet
		if !won {
			trueBet *= -1
		}

		xp.ChangeXP(nil, userID, trueBet, xp.XPS_GAME)

		col := `lost`
		if won {
			col = `won`
		}

		if !db.Exists(`gambling`, `id = $1 AND game = $2`, userID, game) {
			db.Exec(`INSERT INTO gambling (id, game, `+col+`) VALUES ($1, $2, $3)`, userID, game, bet)
		} else {
			db.Exec(`UPDATES gambling SET `+col+` = `+col+`+`+fmt.Sprint(bet)+` WHERE id = $1`, userID)
		}
	}

	ActivityStore.Remove(userID)
}
