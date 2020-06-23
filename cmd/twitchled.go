package main

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
	"github.com/racerxdl/twitchled/twitch"
	"github.com/racerxdl/twitchled/wimatrix"
	"golang.org/x/image/colornames"
	"time"
)

var log = slog.Scope("TwitchLED")
var cfg config.MQTTConfig
var mqttClient mqtt.Client
var ev EventBus.Bus

func OnReward(reward twitch.RedemptionData) {
	log.Debug("User %s rewarded %s", reward.User.DisplayName, reward.Reward.Title)
	switch reward.Reward.Title {
	case config.GetConfig().RewardTitle:
		log.Info("User %s sent %s", reward.User.DisplayName, reward.UserInput)
		ev.Publish(wimatrix.EvNewMsg, fmt.Sprintf("%s by %s", reward.UserInput, reward.User.DisplayName))
	case config.GetConfig().LightRewardTitle:
		log.Info("User %s toggled the light", reward.User.DisplayName)
		ev.Publish(wimatrix.EvSetLight)
	}
}

func main() {
	config.LoadConfig()
	cfg = config.GetConfig()
	log.Info("Connecting to Device %s", cfg.DeviceName)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:1883", cfg.Host))
	opts.SetUsername(cfg.User)
	opts.SetPassword(cfg.Pass)

	mqttClient = mqtt.NewClient(opts)
	log.Debug("Connecting to MQTT at %s", cfg.Host)
	if !mqttClient.Connect().WaitTimeout(time.Second * 5) {
		log.Fatal("Cannot connect to MQTT")
	}

	ev = EventBus.New()

	led := wimatrix.MakeWiiMatrix(cfg.DeviceName, mqttClient, ev)

	led.Start()

	defer led.Stop()

	ev.Publish(wimatrix.EvSetSpeed, int(20))
	ev.Publish(wimatrix.EvNewMode, wimatrix.ModeBackgroundStringDisplay)
	ev.Publish(wimatrix.EvSetTextColor, colornames.Red)
	ev.Publish(wimatrix.EvSetBgColor, colornames.Darkblue)
	ev.Publish(wimatrix.EvSetBgBrightness, float32(0.01))
	ev.Publish(wimatrix.EvSetTextBrightness, float32(0.1))
	ev.Publish(wimatrix.EvNewMsg, "LIVE ON!")

	channelId, err := twitch.GetChannelId()

	if err != nil {
		log.Fatal("Error getting channel id: %s", err)
	}

	channelName, _ := twitch.GetChannelName()

	log.Info("Channel ID is %s and name is %s", channelId, channelName)

	mon := twitch.MakeMonitor(channelId)

	mon.SetCB(OnReward)

	err = mon.Start()
	if err != nil {
		log.Fatal("Error creating monitor: %s", err)
	}

	defer mon.Stop()

	token, _ := twitch.GetAccessToken()

	chat, err := twitch.MakeChat("rxdlbot", "racerxdl", token.AccessToken)

	if err != nil {
		log.Fatal("Error starting chat: %s", err)
	}

	chat.SendMessage("BOT ON!!!")

	msgTimer := time.NewTicker(time.Minute * 5)
	defer msgTimer.Stop()

	log.Info("Waiting messages")
	for {
		select {
		case e := <-chat.Events:
			switch e.GetType() {
			case twitch.EventMessage:
				ParseChat(chat, e.GetData().(*twitch.MessageEventData))
			case twitch.EventError:
				er := e.GetData().(*twitch.ErrorEventData)
				log.Error(er.Error())
				break
			case twitch.EventLoginError:
				er := e.GetData().(*twitch.LoginEventData)
				log.Error(er.Message)
				break
			case twitch.EventLoginSuccess:
				log.Info("Logged in into Twitch Chat")
			}
		case <-msgTimer.C:
			ev.Publish(wimatrix.EvSetSpeed, int(20))

		}
	}

	mqttClient.Disconnect(0)
}
