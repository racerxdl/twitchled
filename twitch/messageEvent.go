package twitch

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

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
	Tags      map[string]string
	Badges    map[string]string
	ExtraData interface{}
	timestamp time.Time
}

func (l *MessageEventData) build() {
	l.Badges = make(map[string]string)
	if badges, ok := l.Tags["badges"]; ok {
		// badges:broadcaster/1,subscriber/0,premium/1
		b := strings.Split(badges, ",")
		for _, v := range b {
			if strings.Contains(v, "/") {
				v2 := strings.Split(v, "/")
				l.Badges[v2[0]] = v2[1]
			} else {
				l.Badges[v] = ""
			}
		}
	}
}

func (l *MessageEventData) IsModerator() bool {
	t, ok := l.Tags["mod"]
	return ok && t == "1"
}

func (l *MessageEventData) IsSubscriber() bool {
	_, ok := l.Badges["subscriber"]

	return ok
}

func (l *MessageEventData) SubscriberMonths() int {
	nv, ok := l.Badges["subscriber"]

	if !ok {
		return 0
	}

	i, _ := strconv.ParseInt(nv, 10, 32)

	return int(i)
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
		"tags":      l.Tags,
	}
}

func (l *MessageEventData) AsJson() string {
	s, _ := json.Marshal(l.AsMap())
	return string(s)
}

func (l *MessageEventData) Timestamp() time.Time {
	return l.timestamp
}

func MakeMessageEventData(source SourceType, username, message, picture string, tags map[string]string, extraData interface{}) *MessageEventData {
	med := &MessageEventData{
		Source:    source,
		Username:  username,
		Message:   message,
		ExtraData: extraData,
		Picture:   picture,
		Tags:      tags,
		timestamp: time.Now(),
	}

	med.build()

	return med
}
