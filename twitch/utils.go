package twitch

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/nicklaw5/helix/v2"
	"github.com/shadiestgoat/log"
)

func isAuthed(u *twitch.User) bool {
	return u.Badges["broadcaster"] == 1 || u.Badges["moderator"] == 1
}

func parseMention(str string) string {
	if str == "" {
		return ""
	}
	if str[0] == '@' {
		str = str[1:]
	}

	return strings.ToLower(str)
}

func userID(name string) string {
	resp, err := helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{
			name,
		},
	})

	if err != nil || len(resp.Data.Users) == 0 {
		return ""
	}

	return resp.Data.Users[0].ID
}

func logError(err error, resp *helix.ResponseCommon, ctx string) bool {
	if err != nil || resp.ErrorMessage != "" {
		msg := "<empty>"
		if resp.ErrorMessage != "" {
			msg = resp.ErrorMessage
		}

		log.Error("While %s: %v %s", ctx, err, msg)
		return true
	}
	
	return false
}
