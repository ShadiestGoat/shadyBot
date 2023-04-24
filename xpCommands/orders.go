package xpCommands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/xp"
)

const (
	ORDER_XP  = "xp"
	ORDER_VC  = "vc"
	ORDER_MSG = "msg"
)

func orderOpt(required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "type",
		Description: "What leaderboard to fetch from",
		Choices: []*discordgo.ApplicationCommandOptionChoice{
			{
				Name:  "Experience",
				Value: ORDER_XP,
			},
			{
				Name:  "Voice Time",
				Value: ORDER_VC,
			},
			{
				Name:  "Message Amount",
				Value: ORDER_MSG,
			},
		},
		Required: required,
	}
}

type Rank struct {
	UserID string

	RankMsg int
	RankVC  int
	RankXP  int

	XP    int
	Level int

	Messages     int
	VoiceMinutes int

	LastActive time.Time
}

func (r *Rank) FillRanks() {
	// Rank shouldn't ever be 0, so this checks if its alr been filled
	if r.RankMsg != 0 {
		return
	}

	r.RankMsg = RankMsg(r.Messages, r.LastActive)
	r.RankVC = RankVC(r.VoiceMinutes, r.LastActive)
	r.RankXP = RankXP(r.Level, r.XP, r.LastActive)
}

func rank(col string, lastUpdate time.Time, v int) int {
	r := 0

	db.QueryRow(`SELECT COUNT(*) FROM xp WHERE `+col+` > $1 OR (`+col+` = $1 AND last_update < $2)`, []any{v, lastUpdate}, &r)

	return r + 1
}

func RankXP(lvl, xp int, lastUpdate time.Time) int {
	r := 0
	db.QueryRow(`SELECT COUNT(*) FROM xp WHERE (lvl > $2) OR (lvl = $2 AND xp > $1) OR (lvl = $2 AND xp = $1 AND last_update < $3)`, []any{xp, lvl, lastUpdate}, &r)

	return r + 1
}

func RankMsg(msgNum int, lastUpdate time.Time) int {
	return rank("msg_num", lastUpdate, msgNum)
}
func RankVC(vcTime int, lastUpdate time.Time) int {
	return rank("vc_time", lastUpdate, vcTime)
}

func FetchUser(userID string) *Rank {
	xp, _ := xp.FetchXP(userID)

	r := &Rank{
		UserID: userID,

		XP:    xp.XP,
		Level: xp.LVL,
	}

	db.QueryRowID(`SELECT vc_time, msg_num, last_update FROM xp WHERE id = $1`, userID, &r.VoiceMinutes, &r.Messages, &r.LastActive)

	return r
}

func orderCols(order string) []string {
	switch order {
	case ORDER_MSG:
		return []string{"msg_num"}
	case ORDER_VC:
		return []string{"vc_time"}
	case ORDER_XP:
		return []string{"lvl", "xp"}
	}

	return []string{}
}

// This can return nil!
func FetchRank(order string, rank int) *Rank {
	r := &Rank{
		RankMsg: 0,
		RankVC:  0,
		RankXP:  0,
	}

	colOrder := orderCols(order)

	switch order {
	case ORDER_MSG:
		r.RankMsg = rank
	case ORDER_VC:
		r.RankVC = rank
	case ORDER_XP:
		r.RankXP = rank
	}

	err := db.QueryRow(`SELECT id, xp, lvl, vc_time, msg_num, last_update FROM xp ORDER BY `+strings.Join(colOrder, " DESC, ")+` DESC, last_update ASC LIMIT 1 OFFSET `+fmt.Sprint(rank-1),
		nil, &r.UserID, &r.XP, &r.Level, &r.VoiceMinutes, &r.Messages, &r.LastActive,
	)

	if err != nil {
		return nil
	}

	return r
}

// Returns row info
func FetchRows(order string, start, stop int) ([]*Rank, error) {
	cols := orderCols(order)

	res, err := db.Query(
		`SELECT id, xp, lvl, vc_time, msg_num, last_update FROM xp ORDER BY `+strings.Join(cols, " DESC, ")+
			` DESC, last_update ASC LIMIT $1 OFFSET $2`, stop-start, start,
	)

	if err != nil {
		return nil, err
	}

	results := []*Rank{}

	for res.Next() {
		resTmp := &Rank{}

		err = res.Scan(&resTmp.UserID, &resTmp.XP, &resTmp.Level, &resTmp.VoiceMinutes, &resTmp.Messages, &resTmp.LastActive)
		if err != nil {
			return nil, err
		}

		results = append(results, resTmp)
	}

	return results, nil
}
