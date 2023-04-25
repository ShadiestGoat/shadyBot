package twitch

import (
	"errors"
	"os"
	"reflect"
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
	userToken *OAuth2
	appToken  *OAuth2

	helixClient  *helix.Client
	ircClient    *twitch.Client
	pubSubClient *twitchpubsub.Client

	OWN_ID   string
	BASE_URL string
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
				for userToken == nil {
					time.Sleep(100 * time.Millisecond)
				}

				helixClient, err = helix.NewClient(&helix.Options{
					ClientID:        config.Twitch.ClientID,
					ClientSecret:    config.Twitch.ClientSecret,
					UserAccessToken: userToken.AccessToken,
					RedirectURI:     BASE_URL,
					ExtensionOpts:   helix.ExtensionOptions{},
				})
				log.FatalIfErr(err, "creating helix client")

				resp, err := helixClient.RequestAppAccessToken(scopes)
				if logError(err, &resp.ResponseCommon, "fetching app access token") {
					log.Fatal("see above")
				}

				helixClient.SetAppAccessToken(resp.Data.AccessToken)

				appToken = &OAuth2{
					AccessToken:  resp.Data.AccessToken,
					RefreshToken: resp.Data.RefreshToken,
					ExpiresIn:    resp.Data.ExpiresIn,
				}

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

				go startIRC()

				log.Success("twitch irc bot setup complete! Adding notifications...")

				if config.Channels.Twitch != "" && config.General.Domain != "localhost" {
					helixClient.SetUserAccessToken("")

					respSubs, err := helixClient.GetEventSubSubscriptions(&helix.EventSubSubscriptionsParams{
						Type: helix.EventSubTypeStreamOnline,
					})

					logError(err, &resp.ResponseCommon, "fetching eventsub subscriptions")

					newSub := &helix.EventSubSubscription{
						Type:    helix.EventSubTypeStreamOnline,
						Version: "1",
						Condition: helix.EventSubCondition{
							BroadcasterUserID: OWN_ID,
						},
						Transport: helix.EventSubTransport{
							Method:   "webhook",
							Callback: BASE_URL + "/live",
							Secret:   config.Twitch.CustomSecret,
						},
					}

					rmID := ""

					log.Debug("New: %#v", newSub)
					
					for _, d := range respSubs.Data.EventSubSubscriptions {
						log.Debug("Old: %#v", d)

						if d.Type != newSub.Type {
							continue
						}
						if d.Version != newSub.Version || !reflect.DeepEqual(d.Condition, newSub.Condition) {
							rmID = d.ID
							break
						}
						if !reflect.DeepEqual(d.Transport, newSub.Transport) {
							rmID = d.ID
							break
						}
					}

					if rmID != "" {
						resp, err := helixClient.RemoveEventSubSubscription(rmID)
						logError(err, &resp.ResponseCommon, "removing eventsub subscription")
						log.Debug("Removing an old and outdates eventsub...")
					}

					resp, err := helixClient.CreateEventSubSubscription(newSub)
					logError(err, &resp.ResponseCommon, "creating evensub subscription")

					helixClient.SetUserAccessToken(userToken.AccessToken)

					if err != nil || resp.ErrorMessage != "" && resp.ErrorMessage != "subscription already exists" {
						msg := ""
						if resp != nil {
							msg = resp.ErrorMessage
						}
						log.Fatal("While creating an event sub for type '%s': %v %s", helix.EventSubTypeStreamOnline, err, msg)
					} else {
						log.Success("Twitch Live notifications ready!")
					}
				}

				go setupPubSub()
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
