package twitch

type SourceType string

const (
	SourceTwitch SourceType = "TWITCH"
)

func (st SourceType) String() string {
	return string(st)
}

type MessageEventData struct {
	Source    SourceType
	Username  string
	Message   string
	Picture   string
	ExtraData interface{}
}

func (l *MessageEventData) GetType() EventType {
	return EventMessage
}

func (l *MessageEventData) GetData() interface{} {
	return l
}

func (l *MessageEventData) GetExtraData() interface{} {
	return l.ExtraData
}

func (l *MessageEventData) GetPicture() string {
	return l.Picture
}

func (l *MessageEventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"type":      l.GetType(),
		"username":  l.Username,
		"source":    l.Source,
		"message":   l.Message,
		"data":      l.GetData(),
		"extraData": l.ExtraData,
		"picture":   l.Picture,
	}
}

func MakeMessageEventData(source SourceType, username, message, picture string, extraData interface{}) *MessageEventData {
	return &MessageEventData{
		Source:    source,
		Username:  username,
		Message:   message,
		ExtraData: extraData,
		Picture:   picture,
	}
}
