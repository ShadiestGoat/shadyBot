package xp

import (
	"fmt"

	"github.com/ShadiestGoat/pronoundb"
	"github.com/shadiestgoat/shadyBot/pronouns"
)

type XPEvent int

const (
	EV_NILL XPEvent = iota
	EV_LVL_UP
	EV_LVL_DOWN
	EV_BONUS
	EV_JACKPOT
)

type XPEventInfo struct {
	Event XPEvent
	// EV_JACKPOT => the jackpot amount
	// EV_BONUS => xp bonus
	// EV_LVL_(DOWN|UP) => the new level
	IntInfo int
	UserID  string
}

func (ev XPEventInfo) String() string {
	switch ev.Event {
	case EV_LVL_UP:
		return fmt.Sprintf("Woohoo! 🎉 <@%v> leveled up to %v! 🎉", ev.UserID, ev.IntInfo)
	case EV_LVL_DOWN:
		pronoun, err := pronouns.Discord(ev.UserID)
		if err != nil {
			pronoun = pronoundb.PR_UNSPECIFIED
		}

		if pronoun.BestGender() == pronoundb.GPR_AVOID {
			return fmt.Sprintf("Haha <@%v> got ratioed down to level %v", ev.UserID, ev.IntInfo)
		}

		return fmt.Sprintf("Oh no! <@%v> leveled down to %v... %v should cry about it :(", ev.UserID, ev.IntInfo, pronoun.They())
	case EV_BONUS:
		return fmt.Sprintf("> <@%v> found a xp bonus of %v! Woohooo!", ev.UserID, ev.IntInfo)
	case EV_JACKPOT:
		return fmt.Sprintf("> ❗ <@%v> got a jackpot of **%v**! Now go gamble your money away in a real casino!", ev.UserID, ev.IntInfo)
	}

	return ""
}

var XPEventChan = make(chan *XPEventInfo, 5)
