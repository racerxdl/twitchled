package twitch

import (
	"encoding/json"
	"github.com/racerxdl/twitchled/twitch/twitchdata"
	"time"
)

type RewardRedemptionEventData struct {
	ChannelId string
	Data      twitchdata.RedemptionData
	timestamp time.Time
}

func (l *RewardRedemptionEventData) GetType() EventType {
	return EventRewardRedemption
}

func (l *RewardRedemptionEventData) GetData() interface{} {
	return l
}

func (l *RewardRedemptionEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":       l.GetType(),
		"channel_id": l.ChannelId,
		"data":       l.GetData(),
		"timestamp":  l.timestamp.Format(time.RFC3339),
	}
}

func (l *RewardRedemptionEventData) AsJson() string {
	s, _ := json.MarshalIndent(l.AsMap(), "", "  ")
	return string(s)
}

func (l *RewardRedemptionEventData) Timestamp() time.Time {
	return l.timestamp
}

func MakeRewardRedemptionEventData(channel string, data twitchdata.RedemptionData) ChatEvent {
	ned := &RewardRedemptionEventData{
		ChannelId: channel,
		Data:      data,
		timestamp: time.Now(),
	}

	return ned
}
