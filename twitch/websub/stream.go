package websub

import "time"

type streamResult struct {
	Data []stream `json:"data"`
}

type stream struct {
	Id           string    `json:"id"`
	UserId       string    `json:"user_id"`
	UserName     string    `json:"user_name"`
	GameId       string    `json:"game_id"`
	CommunityIds []string  `json:"community_ids"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailUrl string    `json:"thumbnail_url"`
}
