package twitch

import (
	"time"
)

type User struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Logo        string `json:"logo"`
	Bio         string `json:"bio"`
}

type Follower struct {
	CreatedAt     time.Time `json:"created_at"`
	Notifications bool      `json:"notifications"`
	User          User      `json:"user"`
}

func MakeUserFromJSON(data map[string]interface{}) User {
	bio := ""

	if data["bio"] != nil {
		bio = data["bio"].(string)
	}

	return User{
		Id:          data["_id"].(string),
		Name:        data["name"].(string),
		Bio:         bio,
		Logo:        data["logo"].(string),
		DisplayName: data["display_name"].(string),
	}
}

func MakeFollowerFromJSON(data map[string]interface{}) Follower {
	createdAt, _ := time.Parse(time.RFC3339, data["created_at"].(string))

	return Follower{
		CreatedAt:     createdAt,
		Notifications: data["notifications"].(bool),
		User:          MakeUserFromJSON(data["user"].(map[string]interface{})),
	}
}
