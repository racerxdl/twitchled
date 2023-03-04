package websub

import (
	"time"
)

type eventsubCondition struct {
	BroadcasterUserId string `json:"broadcaster_user_id"`
}

type eventsubTransport struct {
	Method   string `json:"method"`
	Callback string `json:"callback"`
}

type eventsubEventData struct {
	UserId               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserId    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUsername  string `json:"broadcaster_user_name"`

	// Event Specific
	Type      string    `json:"type"`
	StartedAt time.Time `json:"started_at"`

	// Channel Update Event
	Title        string `json:"title"`
	Language     string `json:"language"`
	CategoryId   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	IsMature     bool   `json:"is_mature"`
}

type eventsubSubscription struct {
	Id      string `json:"id"`
	Status  string `json:"status"`
	Type    string `json:"type"`
	Version string `json:"version"`
	// Cost      string            `json:"cost"`
	Condition eventsubCondition `json:"condition"`
	Transport eventsubTransport `json:"transport"`
	CreatedAt time.Time         `json:"created_at"`
}

type eventsubResponse struct {
	Challenge    string               `json:"challenge"`
	Subscription eventsubSubscription `json:"subscription"`
	Event        eventsubEventData    `json:"event"`
}
