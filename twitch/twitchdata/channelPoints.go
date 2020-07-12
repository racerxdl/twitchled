package twitchdata

import "time"

const (
	ChannelPointsRewardRedeemed = "reward-redeemed"
)

type RedemptionData struct {
	Id         string     `json:"id"`
	User       User       `json:"user"`
	ChannelId  string     `json:"channel_id"`
	RedeemedAt time.Time  `json:"redeemed_at"`
	Reward     RewardData `json:"reward"`
	UserInput  string     `json:"user_input"`
	Status     string     `json:"status"`
}

type RewardData struct {
	Id                                string `json:"id"`
	ChannelId                         string `json:"channel_id"`
	Title                             string `json:"title"`
	Prompt                            string `json:"prompt"`
	Cost                              int    `json:"cost"`
	IsUserInputRequired               bool   `json:"is_user_input_required"`
	IsSubOnly                         bool   `json:"is_sub_only"`
	Image                             Image  `json:"image"`
	DefaultImage                      Image  `json:"default_image"`
	BackgroundColor                   string `json:"background_color"`
	IsEnabled                         bool   `json:"is_enabled"`
	IsPaused                          bool   `json:"is_paused"`
	IsInStock                         bool   `json:"is_in_stock"`
	ShouldRedemptionsSkipRequestQueue bool   `json:"should_redemptions_skip_request_queue"`
}
