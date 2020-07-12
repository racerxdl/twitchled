package twitch

import (
	"encoding/json"
	"github.com/racerxdl/twitchled/twitch/twitchdata"
	"time"
)

type SubscribeEventData struct {
	ChannelId string
	Data      twitchdata.ChannelSubscribeMessageData
	timestamp time.Time
}

func (e *SubscribeEventData) GetType() EventType {
	return EventSubscribe
}

func (e *SubscribeEventData) GetData() interface{} {
	return e.Data
}

func (e *SubscribeEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":       e.GetType(),
		"data":       e.GetData(),
		"channel_id": e.ChannelId,
		"timestamp":  e.timestamp.Format(time.RFC3339),
	}
}

func (e *SubscribeEventData) AsJson() string {
	s, _ := json.Marshal(e.AsMap())
	return string(s)
}

func (e *SubscribeEventData) Timestamp() time.Time {
	return e.timestamp
}

func MakeSubscribeEventData(channelId string, data twitchdata.ChannelSubscribeMessageData) ChatEvent {
	return &SubscribeEventData{
		Data:      data,
		ChannelId: channelId,
		timestamp: time.Now(),
	}
}
