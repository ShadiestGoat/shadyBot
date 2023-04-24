package warnings

import (
	"time"

	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/snownode"
)

func init() {
	cmdWarn()
	cmdPardon()
	cmdWarnings()
}

// 3 Weeks
const TMP_WARNING_TIME = 3 * 7 * 24 * time.Hour

func tmpWarningMinID() string {
	return snownode.TimeToSnow(time.Now().Add(-1 * TMP_WARNING_TIME))
}

func fetch(col string, userID string) (total, tmp int) {
	err := db.QueryRowID(`SELECT `+col+` FROM warnings WHERE warned_user = $1`, userID, &total)
	if err != nil {
		total, tmp = 0, 0
		return
	}

	minID := tmpWarningMinID()

	err = db.QueryRow(`SELECT `+col+` FROM warnings WHERE warned_user = $1 AND id > $2`, []any{userID, minID}, &tmp)

	if err != nil {
		total, tmp = 0, 0
		return
	}

	return
}

func Amount(userID string) (total, tmp int) {
	return fetch(`COUNT(*)`, userID)
}

func Severity(userID string) (total, tmp int) {
	return fetch(`COALESCE(SUM(severity), 0)`, userID)
}

type Warning struct {
	ID         string
	WarnedUser string
	Severity   int
	Reason     string
}

func Punishment(tmpSeverity int, totSeverity int) *config.WarningPunishment {
	punishment := config.WarningPunishment{}

	for _, lvl := range *config.Warnings.Punishments {
		if lvl.MinWarnings <= tmpSeverity {
			punishment = lvl.Punishment
			break
		}
	}

	if totSeverity >= config.Warnings.AutoBanSeverity {
		punishment.Ban = true
	}

	return &punishment
}
