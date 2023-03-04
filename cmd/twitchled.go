package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/asaskevich/EventBus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
	"github.com/racerxdl/twitchled/discord"
	"github.com/racerxdl/twitchled/openai"
	"github.com/racerxdl/twitchled/twitch"
	"github.com/racerxdl/twitchled/twitch/websub"
	"github.com/racerxdl/twitchled/wimatrix"
)

var log = slog.Scope("TwitchLED")
var cfg config.GeneralConfig
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
	discord.SendMessage("FOLLOW", "", strings.Replace(msg, data.Username, "**"+data.Username+"**", -1))
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
		openai.SetLivestreamTitle(data.Title)
		openai.UpdateContext("live_start", time.Now().String())
		_ = chat.SendMessage(fmt.Sprintf("/me LIVE ON!! %s", data.Title))
		discord.Log("TwitchLED", "", "**LIVE ON** everyone! https://twitch.tv/racerxdl")
	} else {
		openai.UpdateContext("live_end", time.Now().String())
		openai.UpdateContext("last_live", openai.GetContext("livestream_title"))
		_ = chat.SendMessage("/me F")
		_ = chat.SendMessage("/me GOODBYE WORLD")
	}
}

func main() {
	config.LoadConfig()
	cfg = config.GetConfig()

	// discord.SendMessage("TwitchLED", "", "**HUEHUE BEGINS**")
	// defer discord.SendMessage("TwitchLED", "", "**GOODBYE WORLD**")
	log.Info("Connecting to Device %s", cfg.DeviceName)
	openai.UpdateContext("bot_start", time.Now().String())
	openai.UpdateContext("livestream_title", "FPGA e compania")

	// opts := mqtt.NewClientOptions()
	// opts.AddBroker(fmt.Sprintf("tcp://%s:1883", cfg.Host))
	// opts.SetUsername(cfg.User)
	// opts.SetPassword(cfg.Pass)

	// mqttClient = mqtt.NewClient(opts)
	// log.Debug("Connecting to MQTT at %s", cfg.Host)
	// if !mqttClient.Connect().WaitTimeout(time.Second * 5) {
	// 	discord.SendMessage("TwitchLED", "", "Cannot connect to MQTT")
	// 	log.Fatal("Cannot connect to MQTT")
	// }

	ev = EventBus.New()

	// led := wimatrix.MakeWiiMatrix(cfg.DeviceName, mqttClient, ev)

	// led.Start()

	// defer led.Stop()
	token, _ := twitch.GetAccessToken()

	// ev.Publish(wimatrix.EvSetSpeed, int(20))
	// ev.Publish(wimatrix.EvNewMode, wimatrix.ModeBackgroundStringDisplay)
	// ev.Publish(wimatrix.EvSetTextColor, colornames.Red)
	// ev.Publish(wimatrix.EvSetBgColor, colornames.Darkblue)
	// ev.Publish(wimatrix.EvSetBgBrightness, float32(0.01))
	// ev.Publish(wimatrix.EvSetTextBrightness, float32(0.1))
	// ev.Publish(wimatrix.EvNewMsg, "LIVE ON!")

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

	wb := websub.MakeSubber()
	go wb.Start(":7002")

	if config.GetConfig().TwitchCallbackBase != "" {
		log.Info("Twitch Callback base set to %q. Registering for events", config.GetConfig().TwitchCallbackBase)
		wb.ClearWebhooks()
		wb.RegisterFollow(channelId)
		wb.RegisterStreamStatus(channelId)
	}

	chat, err := twitch.MakeChat("racerxdl", "racerxdl", token.AccessToken)

	if err != nil {
		log.Fatal("Error starting chat: %s", err)
	}

	// _ = chat.SendMessage("/me HUEHUE BEGINS")

	// msgTimer := time.NewTicker(time.Minute * 5)
	// defer msgTimer.Stop()

	running := true

	c := make(chan os.Signal, 1)
	stop := make(chan bool, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			running = false
			// _ = chat.SendMessage("/me I'M OUT!!!!")
			log.ErrorDone("CLOSING")
			stop <- true
		}
	}()

	recheckClips := time.NewTicker(time.Second * 5)
	defer recheckClips.Stop()

	recheckToken := time.NewTicker(time.Minute * 5)
	defer recheckToken.Stop()

	lastClip := time.Now().Add(time.Hour * -1)
	clips, _ := twitch.GetClips(channelId, lastClip)
	cachedClips := map[string]struct{}{}

	o, err := os.Open(config.GetCacheFileName())
	if err == nil {
		data, _ := ioutil.ReadAll(o)
		_ = json.Unmarshal(data, &cachedClips)
		_ = o.Close()
	}

	log.Info("There are %d clips pending to cache...", len(clips))

	saveCachedClips := func() {
		o, err := os.Create(config.GetCacheFileName())
		if err != nil {
			log.Error("error creating file cacheclips: %s", err)
			return
		}
		defer o.Close()
		data, _ := json.MarshalIndent(&cachedClips, "", "    ")
		o.Write(data)
	}

	log.Info("Waiting messages")
	for running {
		select {
		case <-recheckToken.C:
			twitch.RefreshToken() // It does not refresh if still valid
		case <-recheckClips.C:
			clips, _ := twitch.GetClips(channelId, lastClip)
			for _, v := range clips {
				if _, ok := cachedClips[v]; !ok {
					cachedClips[v] = struct{}{}
					chat.SendMessage(fmt.Sprintf("New clip: %s", v))
					discord.Clip("ClipBot", "", v)
					saveCachedClips()
					lastClip = time.Now()
				}
			}
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
				running = false // force close
			case twitch.EventLoginError:
				er := e.GetData().(*twitch.LoginEventData)
				log.Error(er.Message)
			case twitch.EventLoginSuccess:
				log.Info("Logged in into Twitch Chat")
			}
		// case <-msgTimer.C:
		// 	ev.Publish(wimatrix.EvSetSpeed, int(20))
		case <-stop:
			log.Info("Closing...")
		}
	}

	// mqttClient.Disconnect(0)
}
