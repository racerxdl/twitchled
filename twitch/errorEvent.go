package twitch

type ErrorEventData struct {
	err error
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

func (l *ErrorEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type": l.GetType(),
		"err":  l.err,
		"data": l.GetData(),
	}
}

func MakeErrorEvent(err error) *ErrorEventData {
	return &ErrorEventData{
		err: err,
	}
}
