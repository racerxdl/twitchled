package twitchdata

import "time"

type BadgeEntitlement struct {
	NewVersion      int `json:"new_version"`
	PreviousVersion int `json:"previous_version"`
}

type Emote struct {
	Start int    `json:"start,omitempty"`
	End   int    `json:"end,omitempty"`
	Id    string `json:"id,omitempty"`
}

// ChannelSubscriberMessage is the content message of a Subscribe Event
type ChannelSubscriberMessage struct {
	// The body of the user-entered resub message. Depending on the type of message, the message body contains different fields
	Message string  `json:"message,omitempty"`
	Emotes  []Emote `json:"emotes,omitempty"`
}

// ChannelSubscribeMessageData is the content of the message by a channel-subscribe-events-v1.XXXXX
// Specified at https://dev.twitch.tv/docs/pubsub#example-channel-subscriptions-event-message
type ChannelSubscribeMessageData struct {
	// 	Login name of the person who subscribed or sent a gift subscription
	UserName string `json:"user_name,omitempty"`
	// Display name of the person who subscribed or sent a gift subscription
	DisplayName string `json:"display_name,omitempty"`
	// Name of the channel that has been subscribed or subgifted
	ChannelName string `json:"channel_name,omitempty"`
	// Subscription Plan ID, values: Prime, 1000, 2000, 3000
	SubPlan string `json:"sub_plan,omitempty"`
	// Channel Specific Subscription Plan Name
	SubPlanName string `json:"sub_plan_name,omitempty"`
	// Cumulative number of tenure months of the subscription
	CumulativeMonths int `json:"cumulative_months,omitempty"`
	// Denotes the user’s most recent (and contiguous) subscription tenure streak in the channel
	StreakMonths int `json:"streak_months,omitempty"`
	// If this sub message was caused by a gift subscription
	IsGift bool `json:"is_gift,omitempty"`
	// Message from the subscriber
	SubMessage ChannelSubscriberMessage `json:"sub_message,omitempty"`
	// User ID of the subscription gift recipient
	RecipientId string `json:"recipient_id,omitempty"`
	// Login name of the subscription gift recipient
	RecipientUserName string `json:"recipient_user_name,omitempty"`
	// Display name of the person who received the subscription gift
	RecipientDisplayName string `json:"recipient_display_name,omitempty"`
	// User ID of the person who subscribed or sent a gift subscription
	UserId string `json:"user_id,omitempty"`
	// 	ID of the channel that has been subscribed or subgifted
	ChannelId string `json:"channel_id,omitempty"`
	// Time when the subscription or gift was completed. RFC 3339 format
	Time time.Time `json:"time,omitempty"`
	// Chat message sent with the cheer.
	ChatMessage string `json:"chat_message,omitempty"`
	// Number of bits used.
	BitsUsed int `json:"bits_used,omitempty"`
	// All time total number of Bits used in the channel by a specified user.
	TotalBitsUsed int `json:"total_bits_used,omitempty"`
	// 	Event type associated with the subscription product, values: sub, resub, subgift, anonsubgift, resubgift, anonresubgift
	Context string `json:"context,omitempty"`
	// Information about a user’s new badge level, if the cheer was not anonymous and the user reached a new badge level with this cheer. Otherwise, null.
	BadgeEntitlement *BadgeEntitlement `json:"badge_entitlement,omitempty"`
}
