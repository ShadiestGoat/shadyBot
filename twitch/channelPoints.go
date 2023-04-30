package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/nicklaw5/helix/v2"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/twitch/pubsub"
)

const custom_cmd_title = "Custom Twitch Command"

func updateChannelPoints(rewardID string, redemptionID string, s string) {
	helixClient.UpdateChannelCustomRewardsRedemptionStatus(&helix.UpdateChannelCustomRewardsRedemptionStatusParams{
		ID:            redemptionID,
		BroadcasterID: config.Twitch.OwnID,
		RewardID:      rewardID,
		Status:        s,
	})
}

// reject channel points
func rejectPoints(rewardID string, redemptionID string) {
	updateChannelPoints(rewardID, redemptionID, "CANCELED")
}

// reject channel points
func acceptPoints(rewardID string, redemptionID string) {
	updateChannelPoints(rewardID, redemptionID, "FULFILLED")
}

func refreshPubSub() {
	pubsub.SetToken(userToken.AccessToken)
	pubsub.Connect()
}

var closeOnConnect = make(chan bool, 5)
var closeOnRedeem = make(chan bool, 5)

func handleReward(data *pubsub.Redemption) {
	if data.RewardID != twitchCustomCmdID {
		return
	}

	v := strings.SplitN(data.UserInput, ": ", 2)
	if len(v) == 2 {
		v[0] = strings.ToLower(strings.TrimSpace(v[0]))
		v[1] = strings.TrimSpace(v[1])
	}

	if len(v) != 2 || len(v[0]) < 3 || len(v[1]) < 3 || strings.Contains(v[0], " ") {
		ircClient.Say(config.Twitch.ChannelName, "@"+data.UserDisplayName+", you have to use the correct format for the command! Remember, a command can't have spaces, and must be at least 3 characters in length!")
		rejectPoints(data.RewardID, data.RedemptionID)
		return
	}

	if db.Exists(`twitch_cmd`, `cmd = $1 AND usr = $2`, v[0], data.UserID) {
		ircClient.Say(config.Twitch.ChannelName, "@"+data.UserDisplayName+", you already have a command with that name!")
		rejectPoints(data.RewardID, data.RedemptionID)
		return
	}

	db.Exec(`INSERT INTO twitch_cmd(cmd, usr, resp) VALUES ($1, $2, $3)`, v[0], data.UserID, v[1])

	acceptPoints(data.RewardID, data.RedemptionID)

	ircClient.Say(config.Twitch.ChannelName, "@"+data.UserDisplayName+", your command is ready!")
}

var twitchCustomCmdID = ""

type channelPointRedemptionResp struct {
	Data       []*helix.ChannelCustomRewardsRedemption `json:"data"`
	Pagination helix.Pagination                        `json:"pagination"`
}

func fetchOldRedemptions(after string) {
	v := url.Values{}
	v.Set("broadcaster_id", config.Twitch.OwnID)
	v.Set("reward_id", twitchCustomCmdID)
	v.Set("status", "UNFULFILLED")
	v.Set("first", "50")
	if after != "" {
		v.Set("after", after)
	}

	urlToUse := "https://api.twitch.tv/helix/channel_points/custom_rewards/redemptions?" + v.Encode()
	log.Debug(urlToUse)

	req, _ := http.NewRequest("GET", urlToUse, nil)
	req.Header.Set("Authorization", "Bearer " + userToken.AccessToken)
	req.Header.Set("Client-Id", config.Twitch.ClientID)

	resp, err := http.DefaultClient.Do(req)
	// d :=
	// helix.Pagination

	if err != nil || resp.StatusCode != 200 {
		status := "???"
		body := ""
		if resp != nil {
			status = fmt.Sprint(resp.StatusCode)
			b, _ := io.ReadAll(resp.Body)
			body = string(b)
		}

		log.Error("While fetching channel point redemptions: %v %s '%s'", err, status, body)
		return
	}

	respBody := channelPointRedemptionResp{}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if log.ErrorIfErr(err, "decoding resp") {
		return
	}

	for _, d := range respBody.Data {
		handleReward(&pubsub.Redemption{
			RedemptionID:           d.ID,
			RewardID:               d.Reward.ID,
			UserID:                 d.UserID,
			UserLogin:              d.UserLogin,
			UserDisplayName:        d.UserName,
			UserInputRequired:      d.Reward.IsUserInputRequired,
			UserInput:              d.UserInput,
			ShouldSkipRequestQueue: d.Reward.ShouldRedemptionsSkipRequestQueue,
			RewardTitle:            d.Reward.Title,
			RewardPrompt:           d.Reward.Prompt,
			Time:                   d.RedeemedAt.Time,
		})
	}

	if respBody.Pagination.Cursor != "" {
		fetchOldRedemptions(respBody.Pagination.Cursor)
	}
}

func setupPubSub() {
	resp, err := helixClient.GetCustomRewards(&helix.GetCustomRewardsParams{
		BroadcasterID: config.Twitch.OwnID,
		OnlyManageableRewards: true,
	})
	logError(err, &resp.ResponseCommon, "fetching custom rewards")
	for _, r := range resp.Data.ChannelCustomRewards {
		if r.Title == custom_cmd_title {
			twitchCustomCmdID = r.ID
			break
		}
	}

	if twitchCustomCmdID == "" {
		log.Debug("Creating the custom command custom reward thingy")
		resp, err := helixClient.CreateCustomReward(&helix.ChannelCustomRewardsParams{
			BroadcasterID:                     config.Twitch.OwnID,
			Title:                             "Custom Twitch Command",
			Prompt:                            "Add a command specifically made for you. Format is \"command: what to respond with\"",
			Cost:                              3_200,
			IsEnabled:                         true,
			BackgroundColor:                   "#0d1117",
			IsUserInputRequired:               true,
			IsMaxPerStreamEnabled:             false,
			IsMaxPerUserPerStreamEnabled:      false,
			IsGlobalCooldownEnabled:           false,
			ShouldRedemptionsSkipRequestQueue: true,
		})

		if logError(err, &resp.ResponseCommon, "creating custom cmd") {
			return
		}

		twitchCustomCmdID = resp.Data.ChannelCustomRewards[0].ID
	} else {
		log.Debug("Using old custom reward for twitch custom command <3")
	}

	// Basically, whenever we reconnect, we should handle any previous stuff that was redeemed
	go func() {
		for {
			select {
			case <-pubsub.OnConnect:
				log.Success("Connected to pubsub!")
				fetchOldRedemptions("")
			case <-closeOnConnect:
				return
			}
		}
	}()

	// Basically, whenever we reconnect, we should handle any previous stuff that was redeemed
	go func() {
		for {
			select {
			case r := <-pubsub.Redeems:
				handleReward(r)
			case <-closeOnRedeem:
				return
			}
		}
	}()

	go refreshPubSub()
}
