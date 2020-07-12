package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/gorilla/websocket"
	"github.com/quan-to/slog"
	"net/url"
	"strings"
	"time"
)

const twitchWssUrl = "wss://pubsub-edge.twitch.tv"

const (
	eventBusMessageByNonceTopic    = "twitchws:onMessageByNonce:%s"
	eventBusResponseMessageTopic   = "twitchws:onResponseMessage"
	eventBusWebsocketEvents        = "twitchws:onMessage"
	eventBusWebsocketChannelEvents = "twitchws:onChannelMessage:%s"
)

var log = slog.Scope("TwitchMonitor")

type Monitor struct {
	channelname string
	conn        *websocket.Conn
	pingTimer   *time.Ticker
	done        chan struct{}
	events      chan ChatEvent
	ev          EventBus.Bus
}

func MakeMonitor(channelName string) *Monitor {
	return &Monitor{
		events:      make(chan ChatEvent, 16),
		channelname: channelName,
		ev:          EventBus.New(),
	}
}

func (m *Monitor) Stop() {
	close(m.done)
	if m.conn != nil {
		m.conn.Close()
	}
	if m.pingTimer != nil {
		m.pingTimer.Stop()
	}
	m.ev.Unsubscribe(eventBusWebsocketEvents, m.onEvent)
}

func (m *Monitor) parseMessage(data []byte) {
	var ok bool
	//log.Debug("Received Message: %s", string(data))
	dataMsg := map[string]interface{}{}

	err := json.Unmarshal(data, &dataMsg)
	if err != nil {
		log.Error("Error parsing message: %s\nData: %s", err, string(data))
		return
	}

	twitchType := dataMsg["type"].(string)

	switch twitchType {
	case "PONG":
		//log.Debug("RECEIVED PONG")
	case "RECONNECT":
		log.Debug("Received Reconnect")
		// TODO
	case "RESPONSE":
		nonce := ""
		if nonceI, ok := dataMsg["nonce"]; ok {
			nonce = nonceI.(string)
		}

		if nonce != "" {
			// If we have a nonce, publish in the messageByNonceTopic
			m.ev.Publish(fmt.Sprintf(eventBusMessageByNonceTopic, nonce), dataMsg)
		}

		m.ev.Publish(eventBusResponseMessageTopic, dataMsg)
	case "MESSAGE":
		dataMsg, ok = dataMsg["data"].(map[string]interface{})
		if !ok {
			log.Error("Expected data field to be json")
			return
		}

		m.parseTwitchMessage(dataMsg)
	}
}

func (m *Monitor) parseTwitchMessage(data map[string]interface{}) {
	if _, ok := data["topic"]; !ok {
		log.Error("Expected topic on twitch message")
		return
	}

	if _, ok := data["message"]; !ok {
		log.Error("Expected message on twitch message")
		return
	}

	topic := data["topic"].(string)
	msg := data["message"].(string)

	channelId, event := channelIdAndEventFromTopic(topic)

	if event == "" {
		log.Warn("Unknown event: %s", topic)
		return
	}

	log.Debug("Received event %s for channel %s", event, channelId)

	var twitchEvent ChatEvent
	var err error

	switch event {
	case topicChannelBitsEventv2:
		twitchEvent, err = makeBitsEvent(channelId, msg)
	case topicChannelSubscribe:
		twitchEvent, err = makeSubscribeEvent(channelId, msg)
	case topicChannelPoints:
		twitchEvent, err = makePointsEvent(channelId, msg)
	default:
		err = fmt.Errorf("unknown event: %s", event)
	}

	if err != nil {
		log.Error("error parsing event %s: %s", event, err)
		return
	}

	// Publish to main event bus
	m.ev.Publish(eventBusWebsocketEvents, twitchEvent)

	// Publish to channel event bus
	m.ev.Publish(fmt.Sprintf(eventBusWebsocketChannelEvents, channelId), twitchEvent)
}

func (m *Monitor) sendPing() {
	msg := map[string]string{
		"type": "PING",
	}

	msgBytes, _ := json.Marshal(msg)
	err := m.conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		log.Error("Error sending PING: %s", err)
	}
}

func (m *Monitor) register() {
	token, err := GetAccessToken()
	if err != nil {
		log.Error("Error getting token: %s", err)
	}

	topicList := make([]string, 0)

	topicList = append(topicList, topicChannelPoints+"."+m.channelname)
	topicList = append(topicList, topicChannelSubscribe+"."+m.channelname)
	topicList = append(topicList, topicChannelBitsEventv2+"."+m.channelname)

	err = m.Register(topicList, token.AccessToken)

	if err != nil {
		log.Error("Error registering to topics: %s", err)
	}
}

func (m *Monitor) loop() {
	log.Debug("Starting Loop")
	for {
		select {
		case <-m.done:
			log.Info("Received done. Closing connections")
			break
		case <-m.pingTimer.C:
			//log.Debug("Sending PING")
			m.sendPing()
		}

	}
	log.Debug("Closing loop")
}

func (m *Monitor) messageLoop() {
	log.Debug("Starting message loop")
	for {
		_, msg, err := m.conn.ReadMessage()
		if err != nil && !strings.Contains(err.Error(), "timeout") {
			log.Error("Error receiving message: %s", err)
			close(m.done)
			return
		}

		m.parseMessage(msg)
	}
	log.Debug("Closing message loop")
}

func (m *Monitor) onEvent(data ChatEvent) {
	// Forward received websocket event to event channel
	m.events <- data
}

func (m *Monitor) Start() error {
	u, err := url.Parse(twitchWssUrl)
	if err != nil {
		return err
	}

	log.Info("Connecting to %s", u.String())

	m.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("Error connecting: %s", err)
		return err
	}

	m.pingTimer = time.NewTicker(time.Second * 5)

	m.done = make(chan struct{})

	m.ev.SubscribeAsync(eventBusWebsocketEvents, m.onEvent, false)

	go m.loop()
	go m.messageLoop()

	m.register()

	return nil
}

func (m *Monitor) Register(events []string, token string) error {
	msgId, msg := makeMessage("LISTEN", map[string]interface{}{
		"topics":     events,
		"auth_token": token,
	})

	data, _ := json.Marshal(msg)

	err := m.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}

	result, err := m.waitForResult(msgId, time.Second*5)

	if err != nil { // Timeout
		return err
	}

	if twerror, ok := result["error"]; ok && twerror != "" {
		return fmt.Errorf("twitch error: %s", twerror)
	}

	// If no error provided by twitch, assume LISTEN OK

	return nil
}

func (m *Monitor) EventChannel() chan ChatEvent {
	return m.events
}

func (m *Monitor) waitForResult(msgId string, timeout time.Duration) (data map[string]interface{}, err error) {
	// Create a buffered channel to receive the result
	res := make(chan map[string]interface{}, 1)

	// Create the callback async and once
	m.ev.SubscribeOnceAsync(fmt.Sprintf(eventBusMessageByNonceTopic, msgId), func(data map[string]interface{}) {
		res <- data
	})

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	select {
	case <-timeoutTimer.C:
		// Timeout
		return nil, fmt.Errorf("timeout")
	case data = <-res:
		// Got result
		return data, nil
	}
}
