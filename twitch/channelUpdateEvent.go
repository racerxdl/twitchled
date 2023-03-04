package twitch

import (
	"encoding/json"
	"time"
)

type ChannelUpdateEventData struct {
	ChannelId string
	Username  string
	UserId    string

	Title        string
	Language     string
	CategoryId   string
	CategoryName string
	IsMature     bool

	timestamp time.Time
}

func (e *ChannelUpdateEventData) GetType() EventType {
	return EventChannelUpdate
}

func (e *ChannelUpdateEventData) GetData() interface{} {
	return e
}

func (e *ChannelUpdateEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":          e.GetType(),
		"username":      e.Username,
		"user_id":       e.UserId,
		"channel_id":    e.ChannelId,
		"timestamp":     e.timestamp.Format(time.RFC3339),
		"title":         e.Title,
		"language":      e.Language,
		"category_id":   e.CategoryId,
		"category_name": e.CategoryName,
		"is_mature":     e.IsMature,
	}
}

func (e *ChannelUpdateEventData) AsJson() string {
	s, _ := json.Marshal(e.AsMap())
	return string(s)
}

func (e *ChannelUpdateEventData) Timestamp() time.Time {
	return e.timestamp
}

func MakeChannelUpdateEventData(channelId, username, userId, title, language, categoryId, categoryName string, isMature bool) ChatEvent {
	return &ChannelUpdateEventData{
		ChannelId:    channelId,
		Username:     username,
		UserId:       userId,
		Title:        title,
		Language:     language,
		CategoryId:   categoryId,
		CategoryName: categoryName,
		IsMature:     isMature,
		timestamp:    time.Now(),
	}
}
