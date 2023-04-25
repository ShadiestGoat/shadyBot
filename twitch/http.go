package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
	"github.com/nicklaw5/helix/v2"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/router"
	"github.com/shadiestgoat/shadyBot/snownode"
)

var scopes = []string{
	"chat:edit", "chat:read", "bits:read",
	"channel:manage:redemptions", "channel:read:redemptions",
	"moderator:manage:shoutouts", "moderator:read:shoutouts",
	"user:manage:whispers", "whispers:read", "whispers:edit",
	"moderator:read:followers",
}

type OAuth2 struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

var refreshing = atomic.Bool{}

func refreshToken() {
	if refreshing.Load() {
		return
	}
	refreshing.Store(true)

	if userToken == nil || helixClient == nil {
		return
	}

	resp, err := helixClient.RefreshUserAccessToken(userToken.RefreshToken)
	if log.ErrorIfErr(err, "refreshing twitch token") {
		return
	}

	*userToken = OAuth2{
		AccessToken:  resp.Data.AccessToken,
		RefreshToken: resp.Data.RefreshToken,
	}

	helixClient.SetUserAccessToken(userToken.AccessToken)

	go refreshIRC()
	go refreshPubSub()

	time.Sleep(2 * time.Second)
	refreshing.Store(false)
}

type eventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

func initHTTP(s *discordgo.Session) {
	r := chi.NewRouter()

	state := snownode.Generate()

	log.PrintWarn(
		"Please login for twitch:\nhttps://id.twitch.tv/oauth2/authorize?response_type=code&client_id=%v&redirect_uri=%v&scope=%s&state=%v",
		config.Twitch.ClientID, BASE_URL, url.QueryEscape(strings.Join(scopes, " ")), state,
	)

	r.Get(`/`, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state || q.Get("code") == "" || userToken != nil {
			w.WriteHeader(400)
			w.Write([]byte(`{"error": "You done fucked up man"}`))
			return
		}

		vals := url.Values{}

		vals.Set("client_id", config.Twitch.ClientID)
		vals.Set("client_secret", config.Twitch.ClientSecret)
		vals.Set("redirect_uri", BASE_URL)
		vals.Set("grant_type", "authorization_code")
		vals.Set("code", q.Get("code"))

		resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", vals)
		if err != nil || resp == nil || resp.StatusCode != 200 || resp.Body == nil {
			status, body := 0, "nil"
			if resp != nil {
				status = resp.StatusCode
				if resp.Body != nil {
					b, _ := io.ReadAll(resp.Body)
					body = string(b)
				}
			}
			log.Warn("Couldn't login twitch: %v, resp == nil: %v, %v, '%v'", err, resp == nil, status, body)
			w.WriteHeader(400)
			w.Write([]byte(`{"error": "You done fucked up man"}`))
			return
		}

		authTMP := &OAuth2{}

		if log.ErrorIfErr(json.NewDecoder(resp.Body).Decode(&authTMP), "login into twitch: bad unmarshal") {
			return
		}

		userToken = authTMP

		log.Success("Twitch Authed!")
	})

	r.HandleFunc(`/live`, func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			return
		}
		
		body, _ := io.ReadAll(r.Body)

		log.Debug("Received live notif info: %s", string(body))

		if !helix.VerifyEventSubNotification(config.Twitch.CustomSecret, r.Header, string(body)) {
			log.Error("Illegal notif from twitch!! " + string(body))
			return
		}

		var vals eventSubNotification

		err := json.Unmarshal(body, &vals)
		if log.ErrorIfErr(err, "parsing twitch notification abt live!!"+string(body)) {
			return
		}

		// if there's a challenge in the request, respond with only the challenge to verify your eventsub.
		if vals.Challenge != "" {
			log.Debug("Challenge resp")
			w.WriteHeader(200)
			w.Write([]byte(vals.Challenge))
			return
		}

		if config.Channels.Twitch == "" {
			log.Debug("No twitch announcement channel")
			return
		}

		var liveEvent helix.EventSubStreamOnlineEvent
		err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&liveEvent)

		if log.ErrorIfErr(err, "parsing twitch notification abt live (2nd layer)"+string(body)) {
			return
		}

		emb := discutils.BaseEmbed
		emb.Title = "I'm live!"
		emb.Image = &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("https://static-cdn.jtvnw.net/previews-ttv/live_user_%s-1920x1080.jpg", config.Twitch.ChannelName),
		}

		emb.URL = "https://twitch.tv/" + config.Twitch.ChannelName

		resp, err := helixClient.GetStreams(&helix.StreamsParams{
			UserIDs: []string{
				OWN_ID,
			},
		})

		if err == nil && len(resp.Data.Streams) >= 1 {
			stream := resp.Data.Streams[0]
			emb.Description = stream.Title
		}

		roleMention := ""

		if config.Twitch.Role != "" {
			roleMention = "\n||<@&" + config.Twitch.Role + ">||"
		}

		_, err = discutils.SendMessage(s, config.Channels.Twitch, &discordgo.MessageSend{
			Content: "I'm live!!!!!\n" + emb.URL + roleMention,
			Embeds: []*discordgo.MessageEmbed{
				&emb,
			},
		})

		log.ErrorIfErr(err, "sending the live msg")

		log.Debug("Streaming rn")
	})

	router.Register("/twitch", r)
}
