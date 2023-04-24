package twitch

import (
	"strings"

	"github.com/nicklaw5/helix/v2"
	twitchpubsub "github.com/pajlada/go-twitch-pubsub"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
)

const (
	REWARD_TWITCH_CMD = "Custom Twitch Command"
)

func updateChannelPoints(d *twitchpubsub.PointsEvent, s string) {
	helixClient.UpdateChannelCustomRewardsRedemptionStatus(&helix.UpdateChannelCustomRewardsRedemptionStatusParams{
		ID:            d.Id,
		BroadcasterID: OWN_ID,
		RewardID:      d.Reward.Id,
		Status:        s,
	})
}

// reject channel points
func rejectPoints(d *twitchpubsub.PointsEvent) {
	updateChannelPoints(d, "CANCELED")
}

// reject channel points
func acceptPoints(d *twitchpubsub.PointsEvent) {
	updateChannelPoints(d, "FULFILLED")
}

func pubSub() {
	pubSubClient = twitchpubsub.NewClient(twitchpubsub.DefaultHost)
	pubSubClient.Listen("channel-points-channel-v1."+OWN_ID, userToken.AccessToken)
	pubSubClient.OnPointsEvent(func(_ string, data *twitchpubsub.PointsEvent) {

		switch data.Reward.Title {
		case REWARD_TWITCH_CMD:
			v := strings.SplitN(data.UserInput, ": ", 2)
			if len(v) == 2 {
				v[0] = strings.ToLower(strings.TrimSpace(v[0]))
				v[1] = strings.ToLower(strings.TrimSpace(v[1]))
			}

			if len(v) != 2 || len(v[0]) < 3 || len(v[1]) < 3 || strings.Contains(v[0], " ") {
				ircClient.Say(config.Twitch.ChannelName, "@"+data.User.DisplayName+", you have to use the correct format for the command! Remember, a command can't have spaces, and must be at least 3 characters in length!")
				rejectPoints(data)
				return
			}

			if db.Exists(`twitch_cmd`, `cmd = $1 AND usr = $2`, v[0], data.User.Id) {
				ircClient.Say(config.Twitch.ChannelName, "@"+data.User.DisplayName+", you already have a command with that name!")
				rejectPoints(data)
				return
			}

			db.Exec(`INSERT INTO twitch_cmd(cmd, usr, resp) VALUES ($1, $2, $3)`, v[0], data.User.Id, v[1])

			acceptPoints(data)

			ircClient.Say(config.Twitch.ChannelName, "@"+data.User.DisplayName+", your command is ready!")
		}
	})

	pubSubClient.Start()
}
