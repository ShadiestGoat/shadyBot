package pubsub

import (
	"time"

	"github.com/nicklaw5/helix/v2"
)

type rawReward struct {
	Time       string          `json:"timestamp"`
	Redemption rawRewardRedeem `json:"redemption"`
}

type rawRewardRedeem struct {
	ID         string                `json:"id"`
	User       helix.User            `json:"user"`
	ChannelID  string                `json:"channel_id"`
	RedeemTime string                `json:"redeemed_at"`
	Reward     rawRewardRedeemReward `json:"reward"`
	Input      string                `json:"user_input"`
}

type rawRewardRedeemReward struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	Title     string `json:"title"`
	Prompt    string `json:"prompt"`
	Cost      int    `json:"cost"`

	IsUserInputRequired bool `json:"is_user_input_required"`
	IsSubOnly           bool `json:"is_sub_only"`
	IsEnabled           bool `json:"is_enabled"`
	IsPaused            bool `json:"is_paused"`
	IsInStock           bool `json:"is_in_stock"`

	MaxPerStream     maxPerStream `json:"max_per_stream"`
	SkipRequestQueue bool         `json:"should_redemptions_skip_request_queue"`
}

type maxPerStream struct {
	IsEnabled    bool `json:"is_enabled"`
	MaxPerStream int  `json:"max_per_stream"`
}

func (r rawReward) Parse() *Redemption {
	t, _ := time.Parse(time.RFC3339Nano, r.Redemption.RedeemTime)

	return &Redemption{
		RedemptionID: r.Redemption.ID,
		RewardID:     r.Redemption.Reward.ID,

		UserID:          r.Redemption.User.ID,
		UserLogin:       r.Redemption.User.Login,
		UserDisplayName: r.Redemption.User.DisplayName,

		UserInputRequired: r.Redemption.Reward.IsUserInputRequired,
		UserInput:         r.Redemption.Input,

		ShouldSkipRequestQueue: r.Redemption.Reward.SkipRequestQueue,

		RewardTitle:  r.Redemption.Reward.Title,
		RewardPrompt: r.Redemption.Reward.Prompt,

		Time: t,
	}
}

type Redemption struct {
	RedemptionID string
	RewardID     string

	UserID          string
	UserLogin       string
	UserDisplayName string

	UserInputRequired bool
	UserInput         string

	ShouldSkipRequestQueue bool

	RewardTitle  string
	RewardPrompt string

	Time time.Time
}
