package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/quan-to/slog"
	"net/url"
	"strings"
	"time"
)

const twitchWssUrl = "wss://pubsub-edge.twitch.tv"

var log = slog.Scope("TwitchMonitor")

type OnReward func(reward RedemptionData)

type Monitor struct {
	channelname string
	conn        *websocket.Conn
	pingTimer   *time.Ticker
	done        chan struct{}
	cb          OnReward
}

func MakeMonitor(channelName string) *Monitor {
	return &Monitor{
		channelname: channelName,
	}
}

func (m *Monitor) SetCB(onReward OnReward) {
	m.cb = onReward
}

func (m *Monitor) Stop() {
	close(m.done)
	if m.conn != nil {
		m.conn.Close()
	}
	if m.pingTimer != nil {
		m.pingTimer.Stop()
	}
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
		log.Debug("RECEIVED PONG")
	case "RECONNECT":
		log.Debug("Received Reconnect")
		// TODO
	case "RESPONSE":
		if msgErr, ok := dataMsg["error"].(string); ok {
			log.Error("Error: %s", msgErr)
		}
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

	if strings.Contains(topic, "channel-points-channel") {
		//log.Info("MSG: %s", msg)
		err := json.Unmarshal([]byte(msg), &data)
		if err != nil {
			log.Error(err)
		}
		rtype := data["type"].(string)
		if rtype != "reward-redeemed" {
			return
		}

		v, _ := json.Marshal(data["data"].(map[string]interface{})["redemption"])

		reward := &RedemptionData{}
		_ = json.Unmarshal(v, reward)
		if m.cb != nil {
			m.cb(*reward)
		}
	}
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

func (m *Monitor) registerForBits() {
	token, err := GetAccessToken()
	if err != nil {
		log.Error("Error getting token: %s", err)
	}

	msg := map[string]interface{}{
		"type": "LISTEN",
		"data": map[string]interface{}{
			"topics":     []string{fmt.Sprintf("channel-points-channel-v1.%s", m.channelname)},
			"auth_token": token.AccessToken,
		},
	}

	data, _ := json.Marshal(msg)

	err = m.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Error("Error sending PING: %s", err)
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
			log.Debug("Sending PING")
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

	go m.loop()
	go m.messageLoop()

	m.registerForBits()

	return nil
}
