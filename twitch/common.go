package twitch

type EventType string

const (
	EventLoginSuccess EventType = "LOGIN_SUCCESS"
	EventLoginError   EventType = "LOGIN_ERROR"
	EventMessage      EventType = "MESSAGE"
	EventError        EventType = "ERROR"
	EventUndefined    EventType = "UNDEFINED"
)

func (st EventType) String() string {
	return string(st)
}

type ChatEvent interface {
	GetType() EventType
	GetData() interface{}
	AsMap() map[string]interface{}
}
