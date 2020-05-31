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
	if reward.Reward.Title == config.GetConfig().RewardTitle {
		log.Info("User %s sent %s", reward.User.DisplayName, reward.UserInput)
		ev.Publish(wimatrix.EvNewMsg, fmt.Sprintf("%s by %s", reward.UserInput, reward.User.DisplayName))
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

	ev := EventBus.New()

	led := wimatrix.MakeWiiMatrix(cfg.DeviceName, mqttClient, ev)

	led.Start()

	defer led.Stop()

	ev.Publish(wimatrix.EvNewMode, wimatrix.ModeStringDisplay)
	ev.Publish(wimatrix.EvSetTextColor, colornames.Red)
	ev.Publish(wimatrix.EvNewMsg, "LIVE ON")

	channelId, err := twitch.GetChannelId()

	if err != nil {
		log.Fatal("Error getting channel id: %s", err)
	}

	log.Info("Channel ID is %s", channelId)

	mon := twitch.MakeMonitor(channelId)

	mon.SetCB(OnReward)

	err = mon.Start()
	if err != nil {
		log.Fatal("Error creating monitor: %s", err)
	}

	defer mon.Stop()

	select {}

	mqttClient.Disconnect(0)
}
