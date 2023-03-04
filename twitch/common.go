package twitch

import "time"

type EventType string

const (
	EventUndefined        EventType = "UNDEFINED"
	EventLoginSuccess     EventType = "LOGIN_SUCCESS"
	EventLoginError       EventType = "LOGIN_ERROR"
	EventMessage          EventType = "MESSAGE"
	EventError            EventType = "ERROR"
	EventBits             EventType = "BITS"
	EventSubscribe        EventType = "SUBSCRIBE"
	EventStreamStatus     EventType = "STREAM_STATUS"
	EventRewardRedemption EventType = "REWARD_REDEMPTION"
	EventFollow           EventType = "FOLLOW"
	EventChannelUpdate    EventType = "CHANNEL_UPDATE"
)

func (st EventType) String() string {
	return string(st)
}

type ChatEvent interface {
	GetType() EventType
	GetData() interface{}
	AsMap() map[string]interface{}
	AsJson() string
	Timestamp() time.Time
}
