package twitch

import (
	"encoding/json"
	"github.com/racerxdl/twitchled/twitch/twitchdata"
	"time"
)

type BitsV2EventData struct {
	ChannelId string
	Data      twitchdata.BitEventsV2
	timestamp time.Time
}

func (e *BitsV2EventData) GetType() EventType {
	return EventBits
}

func (e *BitsV2EventData) GetData() interface{} {
	return e.Data
}

func (e *BitsV2EventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":       e.GetType(),
		"data":       e.GetData(),
		"channel_id": e.ChannelId,
		"timestamp":  e.timestamp.Format(time.RFC3339),
	}
}

func (e *BitsV2EventData) AsJson() string {
	s, _ := json.Marshal(e.AsMap())
	return string(s)
}

func (e *BitsV2EventData) Timestamp() time.Time {
	return e.timestamp
}

func MakeBitsV2EventData(channelId string, data twitchdata.BitEventsV2) ChatEvent {
	return &BitsV2EventData{
		Data:      data,
		ChannelId: channelId,
		timestamp: time.Now(),
	}
}
