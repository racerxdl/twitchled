package websub

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
	"github.com/racerxdl/twitchled/twitch"
)

const eventSubApi = "https://api.twitch.tv/helix/eventsub/subscriptions"

var log = slog.Scope("WebSub")

type Subber interface {
	Start(addr string) error
	GetEvents() chan twitch.ChatEvent
	RegisterFollow(channelId string)
	RegisterStreamStatus(channelId string)
	ClearWebhooks()
}

type subber struct {
	events chan twitch.ChatEvent

	// Current channel info
	title        string
	language     string
	categoryId   string
	categoryName string
	isMature     bool
}

func MakeSubber() Subber {
	return &subber{
		events: make(chan twitch.ChatEvent, 16),
	}
}

func (s *subber) Start(addr string) error {
	r := mux.NewRouter()
	r.HandleFunc("/eventsub", s.handleEventSub)

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: time.Second,
		ReadTimeout:  time.Second,
	}

	log.Info("Twitch Callback base set to %s", config.GetConfig().TwitchCallbackBase)

	return srv.ListenAndServe()
}

func (s *subber) GetEvents() chan twitch.ChatEvent {
	return s.events
}

func (s *subber) RegisterFollow(channelId string) {
	s.registerFollow(channelId)
}

func (s *subber) RegisterStreamStatus(channelId string) {
	s.registerLiveStatus(channelId)
}

func (s *subber) ClearWebhooks() {
	req, _ := http.NewRequest("GET", eventSubApi, nil)

	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitch.GetAppToken()))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("error getting webhook list: %s", err)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	var data = struct {
		Data []eventsubSubscription `json:"data"`
	}{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal("error getting webhook list: %s", err)
	}

	for _, wh := range data.Data {
		s.deleteWebhook(wh.Id)
	}
}

func (s *subber) deleteWebhook(webhookId string) {
	log.Info("Removing webhook %s", webhookId)
	req, _ := http.NewRequest("DELETE", eventSubApi+"?id="+webhookId, nil)

	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitch.GetAppToken()))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("error deleting webhook %s: %s", webhookId, err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Debug("Webhook Delete Response: %s", string(body))
}

func (s *subber) registerLiveStatus(channelId string) {
	log.Debug("Registering Channel Update for %s", channelId)
	s.registerWebhook(channelId, "channel.update", "1")
	log.Debug("Registering Live Webhook Start for %s", channelId)
	s.registerWebhook(channelId, "stream.online", "1")
	log.Debug("Registering Live Webhook End for %s", channelId)
	s.registerWebhook(channelId, "stream.offline", "1")
}

func (s *subber) registerFollow(channelId string) {
	log.Debug("Registering follow webhook for %s", channelId)
	s.registerWebhook(channelId, "channel.follow", "2")
}

func (s *subber) registerWebhook(channelId, eventType, version string) {
	cbUrl := fmt.Sprintf("%s/eventsub", config.GetConfig().TwitchCallbackBase)

	payload := map[string]interface{}{
		"transport": map[string]interface{}{
			"callback": cbUrl,
			"method":   "webhook",
			"secret":   config.GetConfig().TwitchCallSecret,
		},
		"condition": map[string]interface{}{
			"broadcaster_user_id": channelId,
			"moderator_user_id":   channelId,
		},
		"type":    eventType,
		"version": version,
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", eventSubApi, bytes.NewReader(jsonData))

	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitch.GetAppToken()))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("error registering webhook: %s", err)
		go func() {
			time.Sleep(time.Second)
			s.registerWebhook(channelId, eventType, version)
		}()
	}

	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK+1 && res.StatusCode != http.StatusOK+2 {
		log.Error("error registering webhook: Status Code == %d", res.StatusCode)
		log.Error("Body: %s", string(body))
		log.Fatal("Aborting")
	}

	_ = res.Body.Close()
}

func (s *subber) handleEventSub(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("Twitch-Eventsub-Message-Id")
	timestamp := r.Header.Get("Twitch-Eventsub-Message-Timestamp")
	sig := r.Header.Get("Twitch-Eventsub-Message-Signature")
	msgType := r.Header.Get("Twitch-Eventsub-Message-Type")
	subType := r.Header.Get("Twitch-Eventsub-Subscription-Type")

	if id == "" || timestamp == "" || sig == "" || msgType == "" {
		// Ignore
		w.WriteHeader(403)
		_, _ = w.Write([]byte("Yo man..."))
		return
	}

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	if !s.validateHMAC(data, r) {
		log.Debug("received invalid challenge")
		w.WriteHeader(403)
		return
	}

	var result eventsubResponse

	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Error("error parsing eventsub body: %s", err)
		w.WriteHeader(500)
		return
	}

	switch msgType {
	case "webhook_callback_verification":
		log.Info("Received challenge for eventSub -> %s", subType)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(result.Challenge))
		return
	case "revocation":
		w.WriteHeader(200)
		return
	}

	if msgType == "notification" {
		fmt.Println(string(data))
		switch result.Subscription.Type {
		case "channel.follow":
			s.handleFollow(result)
		case "channel.update":
			s.handleChannelUpdate(result)
		case "stream.online":
			s.handleStream(result)
		case "stream.offline":
			s.handleStream(result)
		}
		w.WriteHeader(200)
		return
	}

	w.WriteHeader(500)
}

func (s *subber) validateHMAC(data []byte, r *http.Request) bool {
	id := r.Header.Get("Twitch-Eventsub-Message-Id")
	timestamp := r.Header.Get("Twitch-Eventsub-Message-Timestamp")
	sig := r.Header.Get("Twitch-Eventsub-Message-Signature")

	rdata := append([]byte(id), []byte(timestamp)...)
	rdata = append(rdata, data...)
	secret := []byte(config.GetConfig().TwitchCallSecret)

	calculatedSignature := "sha256=" + hex.EncodeToString(signBody(secret, rdata))

	if calculatedSignature != sig {
		log.Debug("INVALID SIGNATURE. Expected %s got %s", calculatedSignature, sig)
		return false
	}

	return true
}

func (s *subber) handleChannelUpdate(res eventsubResponse) {
	s.title = res.Event.Title
	s.categoryId = res.Event.CategoryId
	s.categoryName = res.Event.CategoryName
	s.isMature = res.Event.IsMature
	s.language = res.Event.Language
	s.events <- twitch.MakeChannelUpdateEventData(res.Event.BroadcasterUserId, res.Event.BroadcasterUserLogin, res.Event.BroadcasterUserId, res.Event.Title, res.Event.Language, res.Event.CategoryId, res.Event.CategoryName, res.Event.IsMature)
}

func (s *subber) handleFollow(res eventsubResponse) {
	s.events <- twitch.MakeFollowEventData(res.Event.BroadcasterUserId, res.Event.UserName, res.Event.UserId)
}

func (s *subber) handleStream(res eventsubResponse) {
	channelId := res.Event.BroadcasterUserId
	if res.Subscription.Type == "stream.offline" {
		s.events <- twitch.MakeStreamStatusEventData(channelId, false, "", "", "", "", "", "", "", "", nil, 0, time.Now())
	} else {
		s.events <- twitch.MakeStreamStatusEventData(channelId, true, res.Subscription.Id, res.Event.BroadcasterUserId, res.Event.UserName, "", s.categoryId, s.title, s.language, "", nil, 0, res.Event.StartedAt)
	}
}
