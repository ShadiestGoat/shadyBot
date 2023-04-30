package twitch

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/nicklaw5/helix/v2"
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
		},
		&initializer.ModuleInfo{
			ShouldLoad: func(c *initializer.InitContext) bool {
				return config.Twitch.ShouldLoad()
			},
			PreHooks: []initutils.Module{
				initializer.MOD_HTTP,
			},
		},
		initializer.MOD_DISCORD,
	)

	initializer.Register(initializer.MOD_TWITCH_LIVE, func(c *initializer.InitContext) {
		go func() {
			for userToken == nil {
				time.Sleep(100 * time.Millisecond)
			}

			log.Debug("Starting to actually setup twitch...")
			log.Debug("The user access token expires in %d seconds...", userToken.ExpiresIn)

			var err error

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

			log.Debug("Setup helix! Woohoo!!")

			config.Twitch.OwnID = userID(config.Twitch.ChannelName)

			if config.Twitch.OwnID == "" {
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

			go setupIRC()
			go setupPubSub()

			helixClient.SetUserAccessToken("")
	
			respSubs, err := helixClient.GetEventSubSubscriptions(&helix.EventSubSubscriptionsParams{
				Type: helix.EventSubTypeStreamOnline,
			})
	
			logError(err, &respSubs.ResponseCommon, "fetching eventsub subscriptions")
	
			newSub := &helix.EventSubSubscription{
				Type:    helix.EventSubTypeStreamOnline,
				Version: "1",
				Condition: helix.EventSubCondition{
					BroadcasterUserID: config.Twitch.OwnID,
				},
				Transport: helix.EventSubTransport{
					Method:   "webhook",
					Callback: BASE_URL + "/live",
				},
			}
	
			rmID := []string{}
	
			log.Debug("New: %#v", newSub)
			log.Debug("Resp: %#v", respSubs.Data)

			thereIsGood := false
			
			for _, d := range respSubs.Data.EventSubSubscriptions {
				log.Debug("Old: %#v", d)
								
				if d.Type != newSub.Type {
					continue
				}
				if d.Version != newSub.Version || !reflect.DeepEqual(d.Condition, newSub.Condition) {
					rmID = append(rmID, d.ID)
					continue
				}
				if !reflect.DeepEqual(d.Transport, newSub.Transport) {
					rmID = append(rmID, d.ID)
					continue
				}
				if d.Status == "webhook_callback_verification_failed" {
					rmID = append(rmID, d.ID)
					continue
				}

				thereIsGood = true
			}
	
			if len(rmID) != 0 {
				for _, id := range rmID {
					resp, err := helixClient.RemoveEventSubSubscription(id)
					logError(err, &resp.ResponseCommon, "removing eventsub subscription")
					log.Debug("Removing an old and outdates eventsub...")
				}
			}

			if !thereIsGood {
				newSub.Transport.Secret = config.Twitch.CustomSecret
		
				respNewSub, err := helixClient.CreateEventSubSubscription(newSub)
				logError(err, &respNewSub.ResponseCommon, "creating evensub subscription")
				if err != nil || respNewSub.ErrorMessage != "" && respNewSub.ErrorMessage != "subscription already exists" {
					msg := ""
					if respNewSub != nil {
						msg = respNewSub.ErrorMessage
					}
					log.Fatal("While creating an event sub for type '%s': %v %s", helix.EventSubTypeStreamOnline, err, msg)
				}
			} else {
				log.Debug("Using the old cb, since its still good")
			}
	
			helixClient.SetUserAccessToken(userToken.AccessToken)
	
			log.Success("Twitch Live notifications ready!")
		}()
	}, &initializer.ModuleInfo{
		ShouldLoad: func(c *initializer.InitContext) bool {
			return config.Twitch.ShouldLoad() && config.General.Domain != "localhost"
		},
	}, initializer.MOD_HTTP)
}
