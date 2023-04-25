package twitch

import (
	"fmt"
	"strings"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/nicklaw5/helix/v2"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/pronouns"
)

type TwitchCommand struct {
	Auth bool
	Exec func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string
}

// discord cmd is auto added with the init
var baseCommands = map[string]*TwitchCommand{
	"os": {
		Exec: func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string {
			return "I use arch btw"
		},
	},
	"donate": {
		Exec: func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string {
			return "You can donate at donate.shadygoat.eu, and I will read the message (ps: don't forget to log in with discord to get rewards!)"
		},
	},
	"help": {
		Exec: func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string {
			custom := []string{}

			rows, _ := db.Query(`SELECT cmd FROM twitch_cmd WHERE usr = $1`, ctx.User.ID)

			for rows.Next() {
				cmd := ""

				rows.Scan(&cmd)

				custom = append(custom, "!"+cmd)
			}

			v := "Basic commands are !discord, !donate, !os, !lurk, !pronouns (user)"

			if len(custom) != 0 {
				v += ". You also have custom commands: " + strings.Join(custom, ", ")
			}

			return v
		},
	},
	"so": {
		Auth: true,
		Exec: func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string {
			if len(args) == 0 {
				return ""
			}
			streamer := parseMention(args[0])

			otherStreamer := userID(streamer)

			if otherStreamer == "" {
				return "This person doesn't exist!"
			}

			soResp, err := helixClient.SendShoutout(&helix.SendShoutoutParams{
				FromBroadcasterID: OWN_ID,
				ToBroadcasterID:   otherStreamer,
			})

			logError(err, &soResp.ResponseCommon, "sending twitch so")

			if sp, ok := config.Twitch.CustomSO[streamer]; ok {
				return sp
			}

			return fmt.Sprintf("Hey check out https://twitch.tv/%s, they do really cool stuff!", streamer)
		},
	},

	"lurk": {
		Exec: func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string {
			return "Thank you for the lurk!"
		},
	},

	"pronouns": {
		Exec: func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string {
			usr := ""
			if len(args) != 0 {
				usr = parseMention(args[0])
			}

			if usr == "" || usr == config.Twitch.ChannelName {
				return "My pronouns are she/they, meaning you should refer to me as such. If it's your first time here, or you make an honest mistake, I don't mind it! Just make sure to use the right ones in the future ^^"
			}

			id := userID(usr)

			if id == "" {
				return "Couldn't find this user!"
			}

			pr, err := pronouns.Twitch(id)

			if err != nil {
				return "Sorry, couldn't fetch this user :("
			}

			return pronouns.Explain(pr)
		},
	},
}

func refreshIRC() {
	ircClient.Disconnect()
	ircClient.SetIRCToken("oauth:" + userToken.AccessToken)
	ircClient.Join(config.Twitch.ChannelName)
	log.ErrorIfErr(ircClient.Connect(), "running irc client")
	go refreshToken()
}

func startIRC() {
	log.Debug("Connecting twitch irc oauth...")
	ircClient = twitch.NewClient(config.Twitch.AppName, "oauth:"+userToken.AccessToken)

	ircClient.Join(config.Twitch.ChannelName)

	ircClient.OnConnect(func() {
		log.Success("IRC Loaded <3")
	})

	ircClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		content := message.Message
		if len(content) == 0 {
			return
		}
		if content[0] != '!' {
			return
		}

		args := strings.Split(content[1:], " ")
		command := args[0]
		args = args[1:]
		command = strings.ToLower(command)

		var resp string

		if cmd, ok := baseCommands[command]; ok {
			if cmd.Auth && !isAuthed(&message.User) {
				return
			}

			resp = cmd.Exec(ircClient, &message, args...)
			// prevent ppl from doing infinite loops! This only works if the bot's id is also my id!!!
		} else if message.User.ID != OWN_ID {
			db.QueryRow(`SELECT resp FROM twitch_cmd WHERE cmd = $1 AND usr = $2`, []any{command, message.User.ID}, &resp)
		}

		if resp != "" {
			ircClient.Reply(config.Twitch.ChannelName, message.ID, resp)
		}
	})

	log.ErrorIfErr(ircClient.Connect(), "running irc client")
	go refreshToken()
}
