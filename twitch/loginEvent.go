package twitch

type LoginEventData struct {
	eventType EventType
	Message   string
}

func (l *LoginEventData) GetType() EventType {
	return l.eventType
}

func (l *LoginEventData) GetData() interface{} {
	return l
}

func (l *LoginEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":    l.GetType(),
		"message": l.Message,
		"data":    l.GetData(),
	}
}

func MakeLoginEvent(success bool, message string) *LoginEventData {
	eventType := EventLoginSuccess
	if !success {
		eventType = EventLoginError
	}
	return &LoginEventData{
		eventType: eventType,
		Message:   message,
	}
}
