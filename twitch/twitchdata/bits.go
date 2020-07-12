package twitchdata

import "time"

type BitEventsData struct {
	// Login name of the person who used the Bits - if the cheer was not anonymous. Null if anonymous
	UserName string `json:"user_name"`
	// User ID of the person who used the Bits - if the cheer was not anonymous. Null if anonymous.
	UserId string `json:"user_id"`
	// ID of the channel in which Bits were used.
	ChannelId string `json:"channel_id"`
	// Time when the Bits were used. RFC 3339 format.
	Time time.Time `json:"time"`
	// Chat message sent with the cheer.
	ChatMessage string `json:"chat_message"`
	// Number of bits used.
	BitsUsed int `json:"bits_used"`
	// All time total number of Bits used in the channel by a specified user.
	TotalBitsUsed int `json:"total_bits_used"`
	// 	Event type associated with this use of Bits.
	Context string `json:"context"`
	// Information about a user’s new badge level, if the cheer was not anonymous and the user reached a new badge level with this cheer. Otherwise, null.
	BadgeEntitlement BadgeEntitlement `json:"badge_entitlement"`
}

type BitEventsV2 struct {
	// Data of the bit events
	Data BitEventsData `json:"data"`
	// Message version
	Version string `json:"version"`
	// The type of object contained in the data field.
	MessageType string `json:"message_type"`
	// Message ID.
	MessageId string `json:"message_id"`
	// Whether or not the event was anonymous.
	IsAnonymous bool `json:"is_anonymous"`
}
