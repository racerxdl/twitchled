package twitch

import (
	"encoding/json"
	"time"
)

type LoginEventData struct {
	eventType EventType
	Message   string
	timestamp time.Time
}

func (l *LoginEventData) GetType() EventType {
	return l.eventType
}

func (l *LoginEventData) GetData() interface{} {
	return l
}

func (l *LoginEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":      l.GetType(),
		"message":   l.Message,
		"data":      l.GetData(),
		"timestamp": l.timestamp.Format(time.RFC3339),
	}
}

func (l *LoginEventData) AsJson() string {
	s, _ := json.Marshal(l)
	return string(s)
}

func (l *LoginEventData) Timestamp() time.Time {
	return l.timestamp
}

func MakeLoginEvent(success bool, message string) *LoginEventData {
	eventType := EventLoginSuccess
	if !success {
		eventType = EventLoginError
	}
	return &LoginEventData{
		eventType: eventType,
		Message:   message,
		timestamp: time.Now(),
	}
}
