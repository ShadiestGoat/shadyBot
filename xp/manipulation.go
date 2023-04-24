package xp

import (
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func alterXP(u *XPUser, d int) XPEvent {
	u.XP += d
	ev := EV_NILL

	if d < 0 {
		for u.XP < 0 {
			u.XP += LevelUpRequirement(u.LVL - 1)
			u.LVL--
			ev = EV_LVL_DOWN
		}
	} else {
		for u.XP >= LevelUpRequirement(u.LVL) {
			u.XP -= LevelUpRequirement(u.LVL)
			u.LVL++

			ev = EV_LVL_UP
		}
	}

	return ev
}

type XPSource int

const (
	XPS_NILL XPSource = iota
	XPS_VC
	XPS_MSG
	XPS_GAME
	XPS_CMD
)

func (s XPSource) info() (doMultiplier, doBonus bool) {
	switch s {
	case XPS_NILL, XPS_CMD, XPS_GAME:
		doMultiplier, doBonus = false, false
	}

	return
}

func givenXP(min, max int, factor float64) int {
	return int(math.Round(factor * float64(utils.RandInt(min, max))))
}

// s may be nil, but members will not be fetched (and as such, bonuses/mult don't get applied)
func ChangeXP(s *discordgo.Session, userID string, xpDelta int, source XPSource) *XPUser {
	xpUser, err := FetchXP(userID)
	if err != nil {
		return nil
	}

	if source == XPS_NILL || xpDelta == 0 {
		return xpUser
	}

	shouldMultiply, shouldBonus := source.info()

	if s != nil && shouldMultiply {
		xpDelta = int(math.Round(float64(xpDelta) * multiplier(s, userID)))
	}

	var mem *discordgo.Member

	if s != nil {
		mem = discutils.GetMember(s, config.Discord.GuildID, userID)
	}

	events := []*XPEventInfo{}

	if shouldBonus {
		xpBonus := bonus(mem, xpUser.LVL)

		if xpBonus != 0 {
			xpDelta += xpBonus

			events = append(events, &XPEventInfo{
				Event:   EV_BONUS,
				IntInfo: xpBonus,
				UserID:  userID,
			})
		}
	}

	ev := alterXP(xpUser, xpDelta)

	if ev != EV_NILL {
		events = append(events, &XPEventInfo{
			Event:   ev,
			IntInfo: xpUser.LVL,
			UserID:  userID,
		})
	}

	col := ""

	switch source {
	case XPS_MSG:
		col = "msg_num"
	case XPS_VC:
		col = "vc_time"
	}

	cols := [2]string{}

	if col != "" {
		cols = [2]string{"," + col, "," + col + "+1"}
	}

	sql := fmt.Sprintf(`UPDATE xp SET (xp,lvl%s) = ($1,$2%s) WHERE id = $3`, cols[0], cols[1])

	_, err = db.Exec(sql, xpUser.XP, xpUser.LVL, userID)

	if err == nil {
		for _, ev := range events {
			XPEventChan <- ev
		}
	}

	return xpUser
}
