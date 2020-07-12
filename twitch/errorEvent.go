package twitch

import (
	"encoding/json"
	"time"
)

type ErrorEventData struct {
	err       error
	timestamp time.Time
}

func (l *ErrorEventData) GetType() EventType {
	return EventError
}

func (l *ErrorEventData) GetData() interface{} {
	return l
}

func (l *ErrorEventData) Error() string {
	return l.err.Error()
}

func (l *ErrorEventData) RawError() error {
	return l.err
}

func (l *ErrorEventData) Timestamp() time.Time {
	return l.timestamp
}

func (l *ErrorEventData) AsJson() string {
	s, _ := json.Marshal(l.AsMap())
	return string(s)
}

func (l *ErrorEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type": l.GetType(),
		"err":  l.err,
		"data": l.GetData(),
	}
}

func MakeErrorEvent(err error) ChatEvent {
	return &ErrorEventData{
		err:       err,
		timestamp: time.Now(),
	}
}
