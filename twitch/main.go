package twitch

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/nicklaw5/helix/v2"
	twitchpubsub "github.com/pajlada/go-twitch-pubsub"
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/initializer"
)

var (
	auth         *OAuth2
	helixClient  *helix.Client
	ircClient    *twitch.Client
	pubSubClient *twitchpubsub.Client
	OWN_ID       string
	BASE_URL     string
)

func init() {
	var stopper = make(chan bool)

	initializer.Register(
		initializer.MOD_TWITCH,
		func(c *initializer.InitContext) {
			BASE_URL = "https://" + config.General.Domain

			if config.General.Domain == "localhost" {
				BASE_URL = "http://localhost:" + config.General.Port
			}

			BASE_URL += "/twitch"

			b, err := os.ReadFile("resources/twitch/periodic.md")
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					log.ErrorIfErr(err, "reading periodic twitch text")
				}
			}

			tmpPeriodic := strings.Split(string(b), "\n")

			for _, t := range tmpPeriodic {
				t = strings.TrimSpace(t)
				if t == "" {
					continue
				}

				periodicMessages = append(periodicMessages, t)
			}

			go periodicLaunch(stopper)

			initHTTP(c.Discord)

			go func() {
				for auth == nil {
					time.Sleep(100 * time.Millisecond)
				}

				log.Debug("%#v", auth)

				helixClient, err = helix.NewClient(&helix.Options{
					ClientID:       config.Twitch.ClientID,
					ClientSecret:   config.Twitch.ClientSecret,
					AppAccessToken: auth.AccessToken,
					RedirectURI:    BASE_URL,
					APIBaseURL:     "",
					ExtensionOpts:  helix.ExtensionOptions{},
				})

				OWN_ID = userID(config.Twitch.ChannelName)

				if OWN_ID == "" {
					log.Fatal("Could not fetch our own twitch ID. Are you sure the twitch channel name is correct?")
				}

				log.FatalIfErr(err, "Setting up a helix client")

				if config.Discord.InviteURL != "" {
					baseCommands["discord"] = &TwitchCommand{
						Exec: func(c *twitch.Client, ctx *twitch.PrivateMessage, args ...string) string {
							return "You can join my discord at " + config.Discord.InviteURL
						},
					}
				}

				log.Success("Setup for helix done, running irc...")

				go twitchBot()

				log.Success("twitch irc bot setup complete! Adding notifications...")

				if config.Channels.Twitch != "" {
					createSubscription(helix.EventSubTypeStreamOnline, helix.EventSubCondition{
						BroadcasterUserID: OWN_ID,
					}, "/live")
				}

				go pubSub()
			}()
		},
		&initializer.ModuleInfo{
			ConfigOpts: []*string{
				&config.General.Domain,
				&config.Twitch.ClientID,
				&config.Twitch.ClientSecret,
				&config.Twitch.AppName,
				&config.Twitch.ChannelName,
				&config.Twitch.CustomSecret,
			},
			PreHooks: []initutils.Module{
				initializer.MOD_HTTP,
			},
		},

		initializer.MOD_DISCORD,
	)
}
