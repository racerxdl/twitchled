package main

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
	"github.com/racerxdl/twitchled/discord"
	"github.com/racerxdl/twitchled/twitch"
	"github.com/racerxdl/twitchled/twitch/websub"
	"github.com/racerxdl/twitchled/wimatrix"
	"golang.org/x/image/colornames"
	"os"
	"os/signal"
	"strings"
	"time"
)

var log = slog.Scope("TwitchLED")
var cfg config.MQTTConfig
var mqttClient mqtt.Client
var ev EventBus.Bus

func OnReward(chat *twitch.Chat, reward *twitch.RewardRedemptionEventData) {
	userRewardName := fmt.Sprintf("REWARD(%s)", reward.Data.Reward.Title)
	userRewardAvatar := reward.Data.Reward.Image.Url4x

	log.Debug("User %s rewarded %s", reward.Data.User.DisplayName, reward.Data.Reward.Title)

	switch reward.Data.Reward.Title {
	case config.GetConfig().RewardTitle:
		msg := fmt.Sprintf("%s by %s", reward.Data.UserInput, reward.Data.User.DisplayName)
		log.Info("User %s sent %s", reward.Data.User.DisplayName, reward.Data.UserInput)
		discord.SendMessage(userRewardName, userRewardAvatar, fmt.Sprintf("Panel from %s", reward.Data.User.DisplayName))
		ev.Publish(wimatrix.EvNewMsg, msg)
		_ = chat.SendMessage(fmt.Sprintf("Panel set to: %s", msg))
	case config.GetConfig().LightRewardTitle:
		log.Info("User %s toggled the light", reward.Data.User.DisplayName)
		discord.SendMessage(userRewardName, userRewardAvatar, fmt.Sprintf("Light toggle from %s", reward.Data.User.DisplayName))
		ev.Publish(wimatrix.EvSetLight)
	case config.GetConfig().CodeReviewRewardTitle:
		log.Info("User %s requested a code review: %s", reward.Data.User.DisplayName, reward.Data.UserInput)
		discord.SendMessage(userRewardName, userRewardAvatar, fmt.Sprintf("@here - Code Review from **%s**: %s", reward.Data.User.DisplayName, reward.Data.UserInput))
	}
}

func OnFollow(chat *twitch.Chat, data *twitch.FollowEventData) {
	msg := fmt.Sprintf("User %s followed", data.Username)
	log.Debug(msg)
	_ = chat.SendMessage(fmt.Sprintf("Thanks %s for the follow!", data.Username))
	_ = chat.SendMessage(fmt.Sprintf("Obrigado %s pelo follow!", data.Username))
	discord.SendMessage("FOLLOW", "", strings.Replace(msg, data.Username, "**"+data.UserId+"**", -1))
}

func OnBits(chat *twitch.Chat, bits *twitch.BitsV2EventData) {
	username := bits.Data.Data.UserName
	if bits.Data.IsAnonymous {
		username = "Anonymous"
	}
	numBits := bits.Data.Data.BitsUsed
	message := bits.Data.Data.ChatMessage
	msg := fmt.Sprintf("User %s send %d bits: %s!", username, numBits, message)
	log.Info(msg)
	ev.Publish(wimatrix.EvNewBits, username, numBits, message)
	_ = chat.SendMessage(fmt.Sprintf("Thanks %s for %d bits!!", username, numBits))
	_ = chat.SendMessage(fmt.Sprintf("Obrigado %s por %d bits!!", username, numBits))
	discord.SendMessage("BITS", "", msg)
}

func OnSub(chat *twitch.Chat, subscribe *twitch.SubscribeEventData) {
	msg := fmt.Sprintf("User %s subscribed for %d months!", subscribe.Data.DisplayName, subscribe.Data.StreakMonths+1)
	log.Info(msg)
	ev.Publish(wimatrix.EvNewSub, subscribe.Data.DisplayName, subscribe.Data.StreakMonths+1)
	_ = chat.SendMessage(fmt.Sprintf("Thanks @%s for %d months subscription!!", subscribe.Data.DisplayName, subscribe.Data.StreakMonths+1))
	_ = chat.SendMessage(fmt.Sprintf("Obrigado @%s pelo sub de %d meses!!", subscribe.Data.DisplayName, subscribe.Data.StreakMonths+1))
	discord.SendMessage("SUBSCRIBE", "", msg)
}

func OnStreamChange(chat *twitch.Chat, data *twitch.StreamStatusEventData) {
	if data.Online {
		_ = chat.SendMessage(fmt.Sprintf("/me LIVE ON MEUS CONSAGRADOS!! %s", data.Title))
		discord.Log("TwitchLED", "", "**LIVE ON** @everyone! https://twitch.tv/racerxdl")
	} else {
		_ = chat.SendMessage("/me F")
		_ = chat.SendMessage("/me GOODBYE WORLD")
	}
}

func main() {
	config.LoadConfig()
	cfg = config.GetConfig()

	discord.SendMessage("TwitchLED", "", "**HUEHUE BEGINS**")
	defer discord.SendMessage("TwitchLED", "", "**GOODBYE WORLD**")
	log.Info("Connecting to Device %s", cfg.DeviceName)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:1883", cfg.Host))
	opts.SetUsername(cfg.User)
	opts.SetPassword(cfg.Pass)

	mqttClient = mqtt.NewClient(opts)
	log.Debug("Connecting to MQTT at %s", cfg.Host)
	if !mqttClient.Connect().WaitTimeout(time.Second * 5) {
		discord.SendMessage("TwitchLED", "", "Cannot connect to MQTT")
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

	_ = chat.SendMessage("/me HUEHUE BEGINS")

	msgTimer := time.NewTicker(time.Minute * 5)
	defer msgTimer.Stop()

	running := true

	c := make(chan os.Signal, 1)
	stop := make(chan bool, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			running = false
			_ = chat.SendMessage("/me FUCK THAT, I'M OUT!!!!")
			log.ErrorDone("CLOSING")
			stop <- true
		}
	}()

	log.Info("Waiting messages")
	for running {
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
				OnReward(chat, e.(*twitch.RewardRedemptionEventData))
			case twitch.EventBits:
				OnBits(chat, e.(*twitch.BitsV2EventData))
			case twitch.EventSubscribe:
				OnSub(chat, e.(*twitch.SubscribeEventData))
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
		case <-stop:
			log.Info("Closing...")
		}
	}

	mqttClient.Disconnect(0)
}
