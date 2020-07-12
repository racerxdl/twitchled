package twitch

import (
	"encoding/json"
	"time"
)

type StreamStatusEventData struct {
	ChannelId    string
	Online       bool
	Id           string
	UserId       string
	UserName     string
	GameId       string
	CommunityIds []string
	Type         string
	Title        string
	ViewerCount  int
	StartedAt    time.Time
	Language     string
	ThumbnailUrl string
	timestamp    time.Time
}

func (e *StreamStatusEventData) GetType() EventType {
	return EventStreamStatus
}

func (e *StreamStatusEventData) GetData() interface{} {
	return e
}

func (e *StreamStatusEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"id":            e.Id,
		"online":        e.Online,
		"user_id":       e.UserId,
		"user_name":     e.UserName,
		"game_id":       e.GameId,
		"community_ids": e.CommunityIds,
		"title":         e.Title,
		"viewer_count":  e.ViewerCount,
		"started_at":    e.StartedAt,
		"language":      e.Language,
		"thumbnail_url": e.ThumbnailUrl,
		"type":          e.GetType(),
		"data":          e.GetData(),
		"channel_id":    e.ChannelId,
		"timestamp":     e.timestamp.Format(time.RFC3339),
	}
}

func (e *StreamStatusEventData) AsJson() string {
	s, _ := json.Marshal(e.AsMap())
	return string(s)
}

func (e *StreamStatusEventData) Timestamp() time.Time {
	return e.timestamp
}

func MakeStreamStatusEventData(channelId string, online bool, id, userId, userName, gameId, stype, title, language, thumbnail string, communityIds []string, viewers int, startedAt time.Time) ChatEvent {
	return &StreamStatusEventData{
		ChannelId:    channelId,
		Online:       online,
		Id:           id,
		UserId:       userId,
		UserName:     userName,
		GameId:       gameId,
		Type:         stype,
		Title:        title,
		Language:     language,
		ThumbnailUrl: thumbnail,
		ViewerCount:  viewers,
		StartedAt:    startedAt,
		CommunityIds: communityIds,
		timestamp:    time.Now(),
	}
}
