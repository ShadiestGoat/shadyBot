package twitch

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/nicklaw5/helix/v2"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
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

func createSubscription(t string, cond helix.EventSubCondition, cbPath string) {
	if config.General.Domain == "localhost" {
		return
	}
	
	resp, err := helixClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:      t,
		Version:   "1",
		Condition: cond,
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: BASE_URL + cbPath,
			Secret:   config.Twitch.CustomSecret,
		},
	})

	if err != nil || resp.ErrorMessage != "" && resp.ErrorMessage != "subscription already exists" {
		msg := ""
		if resp != nil {
			msg = resp.ErrorMessage
		}
		log.Fatal("While creating an event sub for type '%s': %v %s", t, err, msg)
	} else {
		log.Success("Twitch Live notifications ready!")
	}
}
