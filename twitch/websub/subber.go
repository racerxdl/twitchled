package websub

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
	"github.com/racerxdl/twitchled/twitch"
	"io/ioutil"
	"net/http"
	"time"
)

const helixHub = "https://api.twitch.tv/helix/webhooks/hub"
const leaseSeconds = 3600 * 24 * 2 // Two Days

const (
	followsTopic      = "https://api.twitch.tv/helix/users/follows?first=1&to_id=%s"
	streamStatusTopic = "https://api.twitch.tv/helix/streams?user_id=%s"
)

var log = slog.Scope("WebSub")

type Subber interface {
	Start(addr string) error
	GetEvents() chan twitch.ChatEvent
	RegisterFollow(channelId string)
	RegisterStreamStatus(channelId string)
}

type subber struct {
	events chan twitch.ChatEvent
}

func MakeSubber() Subber {
	return &subber{
		events: make(chan twitch.ChatEvent, 16),
	}
}

func (s *subber) Start(addr string) error {
	r := mux.NewRouter()
	r.HandleFunc("/follow", s.handleFollow)
	r.HandleFunc("/stream", s.handleStream)

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
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

func (s *subber) registerLiveStatus(channelId string) {
	payload := map[string]interface{}{
		"hub.callback":      fmt.Sprintf("%s/stream", config.GetConfig().TwitchCallbackBase),
		"hub.mode":          "subscribe",
		"hub.topic":         fmt.Sprintf(streamStatusTopic, channelId),
		"hub.lease_seconds": leaseSeconds,
		"hub.secret":        config.GetConfig().TwitchCallSecret,
	}

	log.Debug("Registering Live Webhook for %s on %s", channelId, payload["hub.callback"])
	s.registerWebhook(payload)
}

func (s *subber) registerFollow(channelId string) {
	payload := map[string]interface{}{
		"hub.callback":      fmt.Sprintf("%s/follow", config.GetConfig().TwitchCallbackBase),
		"hub.mode":          "subscribe",
		"hub.topic":         fmt.Sprintf(followsTopic, channelId),
		"hub.lease_seconds": leaseSeconds,
		"hub.secret":        config.GetConfig().TwitchCallSecret,
	}

	log.Debug("Registering webhook for %s with secret %q", channelId, config.GetConfig().TwitchCallSecret)
	s.registerWebhook(payload)
}

func (s *subber) registerWebhook(data map[string]interface{}) {
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", helixHub, bytes.NewReader(jsonData))

	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitch.GetAppToken()))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("error registering webhook: %s", err)
		go func() {
			time.Sleep(time.Second)
			s.registerWebhook(data)
		}()
	}

	if res.StatusCode != http.StatusOK+2 {
		log.Error("error registering webhook: Status Code == %d", res.StatusCode)
		go func() {
			time.Sleep(time.Second)
			s.registerWebhook(data)
		}()
	}

	_ = res.Body.Close()
}

func (s *subber) handleFollow(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	challenge := q.Get("hub.challenge")

	if challenge != "" {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(challenge))
		log.Info("Received Follow Challenge")

		return
	}

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	sig := r.Header.Get("X-Hub-Signature")
	calculatedSignature := "sha256=" + hex.EncodeToString(signBody([]byte(config.GetConfig().TwitchCallSecret), data))

	if calculatedSignature != sig {
		log.Debug("INVALID SIGNATURE. Expected %s got %s", calculatedSignature, sig)
		return
	}

	res := followResult{}

	err = json.Unmarshal(data, &res)

	if err != nil {
		w.WriteHeader(500)
		errmsg := fmt.Sprintf("error parsing data: %s", err)
		log.Error(errmsg)
		w.Write([]byte(errmsg))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("OK"))

	channelId := getChannelId(r.Header.Get("Link"))

	if len(res.Data) > 0 {
		s.events <- twitch.MakeFollowEventData(channelId, res.Data[0].FromName, res.Data[0].FromId)
		//s.registerFollow(res.Data[0].ToId) // Probably not needed
	}

}

func (s *subber) handleStream(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	challenge := q.Get("hub.challenge")

	if challenge != "" {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(challenge))
		log.Info("Received Stream Challenge")

		return
	}

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	sig := r.Header.Get("X-Hub-Signature")
	calculatedSignature := "sha256=" + hex.EncodeToString(signBody([]byte(config.GetConfig().TwitchCallSecret), data))

	if calculatedSignature != sig {
		log.Debug("INVALID SIGNATURE. Expected %s got %s", calculatedSignature, sig)
		return
	}

	channelId := getChannelId(r.Header.Get("Link"))

	res := streamResult{}

	err = json.Unmarshal(data, &res)

	if err != nil {
		w.WriteHeader(500)
		errmsg := fmt.Sprintf("error parsing data: %s", err)
		log.Error(errmsg)
		w.Write([]byte(errmsg))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("OK"))

	if len(res.Data) > 0 {
		d := res.Data[0]
		s.events <- twitch.MakeStreamStatusEventData(channelId, true, d.Id, d.UserId, d.UserName, d.GameId, d.Type, d.Title, d.Language, d.ThumbnailUrl, d.CommunityIds, d.ViewerCount, d.StartedAt)
	} else {
		s.events <- twitch.MakeStreamStatusEventData(channelId, false, "", "", "", "", "", "", "", "", nil, 0, time.Now())
	}

}
