package twitch

import (
	"encoding/json"
	"time"
)

type FollowEventData struct {
	ChannelId string
	Username  string
	UserId    string
	timestamp time.Time
}

func (e *FollowEventData) GetType() EventType {
	return EventFollow
}

func (e *FollowEventData) GetData() interface{} {
	return e
}

func (e *FollowEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":       e.GetType(),
		"username":   e.Username,
		"user_id":    e.UserId,
		"channel_id": e.ChannelId,
		"timestamp":  e.timestamp.Format(time.RFC3339),
	}
}

func (e *FollowEventData) AsJson() string {
	s, _ := json.Marshal(e.AsMap())
	return string(s)
}

func (e *FollowEventData) Timestamp() time.Time {
	return e.timestamp
}

func MakeFollowEventData(channelId, username, userId string) ChatEvent {
	return &FollowEventData{
		ChannelId: channelId,
		Username:  username,
		UserId:    userId,
		timestamp: time.Now(),
	}
}
