package websub

import "time"

type streamResult struct {
	Data []stream `json:"data"`
}

/*
   {
     "id": "0123456789",
     "user_id": "5678",
     "user_name": "wjdtkdqhs",
     "game_id": "21779",
     "community_ids": [],
     "type": "live",
     "title": "Best Stream Ever",
     "viewer_count": 417,
     "started_at": "2017-12-01T10:09:45Z",
     "language": "en",
     "thumbnail_url": "https://link/to/thumbnail.jpg"
   }
*/
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
