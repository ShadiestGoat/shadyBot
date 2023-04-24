package xp

import (
	"github.com/shadiestgoat/shadyBot/db"
)

type XPUser struct {
	XP     int
	LVL    int
	UserID string
}

func (u XPUser) UpdateSQL() error {
	_, err := db.Exec(`UPDATE xp SET xp = $1, lvl = $2 WHERE id = $3`, u.XP, u.LVL, u.UserID)
	return err
}

func (u XPUser) HasNeededXP(xp int) bool {
	totXP := LvlXP(u.LVL) + u.XP
	return totXP >= xp
}

func FetchXP(userID string) (*XPUser, error) {
	var (
		xp  int
		lvl int
	)

	err := db.QueryRowID(`SELECT xp, lvl FROM xp WHERE id = $1`, userID, &xp, &lvl)

	if err != nil && db.NoRows(err) {
		_, err = db.Exec(`INSERT INTO xp(id) VALUES ($1)`, userID)
	}

	return &XPUser{
		XP:     xp,
		LVL:    lvl,
		UserID: userID,
	}, err
}
