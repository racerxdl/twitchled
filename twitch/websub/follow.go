package websub

import "time"

type followResult struct {
	Data []follower `json:"data"`
}

type follower struct {
	FollowedAt time.Time `json:"followed_at"`
	FromId     string    `json:"from_id"`
	FromName   string    `json:"from_name"`
	ToId       string    `json:"to_id"`
	ToName     string    `json:"to_name"`
}
