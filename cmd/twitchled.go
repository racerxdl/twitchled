package main

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
	"golang.org/x/image/colornames"
	"image/color"
	"time"
)

var log = slog.Scope("TwitchLED")
var cfg config.MQTTConfig
var mqttClient mqtt.Client

func PostMessage(message string, c color.Color) error {
	topic := cfg.DeviceName + "_msg"

	r, g, b, _ := c.RGBA()

	data := map[string]interface{}{
		"msg": message,
		"r":   r,
		"g":   g,
		"b":   b,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	tkn := mqttClient.Publish(topic, 0, false, dataBytes)
	if !tkn.WaitTimeout(time.Second) {
		return tkn.Error()
	}

	return nil
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

	PostMessage("GOSCRIPTO ???", colornames.Magenta)

	mqttClient.Disconnect(0)
}
