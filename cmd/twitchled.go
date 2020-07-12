package main

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
	"github.com/racerxdl/twitchled/twitch"
	"github.com/racerxdl/twitchled/twitch/websub"
	"github.com/racerxdl/twitchled/wimatrix"
	"golang.org/x/image/colornames"
	"time"
)

var log = slog.Scope("TwitchLED")
var cfg config.MQTTConfig
var mqttClient mqtt.Client
var ev EventBus.Bus

func OnReward(chat *twitch.Chat, reward *twitch.RewardRedemptionEventData) {
	log.Debug("User %s rewarded %s", reward.Data.User.DisplayName, reward.Data.Reward.Title)
	switch reward.Data.Reward.Title {
	case config.GetConfig().RewardTitle:
		msg := fmt.Sprintf("%s by %s", reward.Data.UserInput, reward.Data.User.DisplayName)
		log.Info("User %s sent %s", reward.Data.User.DisplayName, reward.Data.UserInput)
		ev.Publish(wimatrix.EvNewMsg, msg)
		chat.SendMessage(fmt.Sprintf("Panel: %s", msg))
	case config.GetConfig().LightRewardTitle:
		log.Info("User %s toggled the light", reward.Data.User.DisplayName)
		ev.Publish(wimatrix.EvSetLight)
	}
}

func OnFollow(chat *twitch.Chat, data *twitch.FollowEventData) {
	log.Debug("User %s followed", data.Username)
	chat.SendMessage(fmt.Sprintf("Thanks %s for the follow!", data.Username))
	chat.SendMessage(fmt.Sprintf("Obrigado %s pelo follow!", data.Username))
}

func OnBits(chat *twitch.Chat, bits *twitch.BitsV2EventData) {
	username := bits.Data.Data.UserName
	if bits.Data.IsAnonymous {
		username = "Anonymous"
	}
	numBits := bits.Data.Data.BitsUsed
	message := bits.Data.Data.ChatMessage

	log.Info("User %s send %d bits: %s!", username, numBits, message)
	ev.Publish(wimatrix.EvNewBits, username, numBits, message)
	chat.SendMessage(fmt.Sprintf("Thanks %s for %d bits!!", username, numBits))
	chat.SendMessage(fmt.Sprintf("Obrigado %s por %d bits!!", username, numBits))
}

func OnSub(chat *twitch.Chat, subscribe *twitch.SubscribeEventData) {
	log.Info("User %s subscribed for %d months!", subscribe.Data.DisplayName, subscribe.Data.StreakMonths)
	ev.Publish(wimatrix.EvNewSub, subscribe.Data.DisplayName, subscribe.Data.StreakMonths)
	chat.SendMessage(fmt.Sprintf("Thanks @%s for %d months subscription!!", subscribe.Data.DisplayName, subscribe.Data.StreakMonths))
	chat.SendMessage(fmt.Sprintf("Obrigado @%s pelo sub de %d meses!!", subscribe.Data.DisplayName, subscribe.Data.StreakMonths))
}

func OnStreamChange(chat *twitch.Chat, data *twitch.StreamStatusEventData) {
	if data.Online {
		chat.SendMessage(fmt.Sprintf("/me LIVE ON MEUS CONSAGRADOS!! %s", data.Title))
	} else {
		chat.SendMessage("/me F")
		chat.SendMessage("/me GOODBYE WORLD")
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

	err = mon.Start()
	if err != nil {
		log.Fatal("Error creating monitor: %s", err)
	}

	defer mon.Stop()

	token, _ := twitch.GetAccessToken()

	wb := websub.MakeSubber()
	go wb.Start(":7002")

	if config.GetConfig().TwitchCallbackBase != "" {
		log.Info("Twitch Callback base set to %s. Registering for events", config.GetConfig().TwitchCallbackBase)
		wb.RegisterFollow(channelId)
		wb.RegisterStreamStatus(channelId)
	}

	chat, err := twitch.MakeChat("racerxdl", channelName, token.AccessToken)

	if err != nil {
		log.Fatal("Error starting chat: %s", err)
	}

	chat.SendMessage("/me HUEHUE BEGINS")

	msgTimer := time.NewTicker(time.Minute * 5)
	defer msgTimer.Stop()

	log.Info("Waiting messages")
	for {
		select {
		case e := <-wb.GetEvents():
			switch e.GetType() {
			case twitch.EventFollow:
				OnFollow(chat, e.GetData().(*twitch.FollowEventData))
			case twitch.EventStreamStatus:
				OnStreamChange(chat, e.GetData().(*twitch.StreamStatusEventData))
			}
		case e := <-mon.EventChannel():
			switch e.GetType() {
			case twitch.EventRewardRedemption:
				OnReward(chat, e.GetData().(*twitch.RewardRedemptionEventData))
			case twitch.EventBits:
				OnBits(chat, e.GetData().(*twitch.BitsV2EventData))
			case twitch.EventSubscribe:
				OnSub(chat, e.GetData().(*twitch.SubscribeEventData))
			}
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
