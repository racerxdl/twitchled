package config

import (
	"encoding/base64"
	"github.com/BurntSushi/toml"
	"github.com/mewkiz/pkg/osutil"
	"github.com/quan-to/slog"
	"os"
	"strings"
)

type MQTTConfig struct {
	Host                  string
	User                  string
	Pass                  string
	DeviceName            string
	TwitchOAuthClient     string
	TwitchOAuthSecret     string
	TwitchTokenData       string
	TwitchAppTokenData    string
	TwitchCallbackBase    string
	RewardTitle           string
	LightRewardTitle      string
	CodeReviewRewardTitle string
	TwitchCallSecret      string
	DiscordBotOutputUrl   string
	DiscordLogOutputUrl   string
	LogIgnoreList         string
}

func IsOnIgnoreList(username string) bool {
	username = strings.ToLower(username)
	ignoreList := strings.Split(config.LogIgnoreList, ",")

	for _, v := range ignoreList {
		v = strings.ToLower(strings.Trim(v, " \r\n"))
		if v == username {
			return true
		}
	}

	return false
}

const configFile = "twitchled.toml"

var config MQTTConfig

var log = slog.Scope("MCP2MQTT")

func GetConfig() MQTTConfig {
	return config
}

func SetTwitchToken(tokenData []byte) {
	config.TwitchTokenData = base64.StdEncoding.EncodeToString(tokenData)
	SaveConfig()
}

func SetTwitchAppTokenData(tokenData []byte) {
	config.TwitchAppTokenData = base64.StdEncoding.EncodeToString(tokenData)
	SaveConfig()
}

func LoadConfig() {
	log.Info("Loading config %s", configFile)
	if !osutil.Exists(configFile) {
		log.Error("Config file %s does not exists.", configFile)
		os.Exit(1)
	}

	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		log.Error("Error decoding file %s: %s", configFile, err)
		os.Exit(1)
	}
}

func SaveConfig() {
	log.Info("Saving config")
	f, err := os.Create(configFile)
	if err != nil {
		log.Fatal("Error opening %s: %s", configFile, err)
	}
	e := toml.NewEncoder(f)
	err = e.Encode(&config)
	if err != nil {
		log.Fatal("Error saving data to %s: %s", configFile, err)
	}
}
