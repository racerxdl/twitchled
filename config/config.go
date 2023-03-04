package config

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mewkiz/pkg/osutil"
	"github.com/quan-to/slog"
)

type GeneralConfig struct {
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
	DiscordClipOutputUrl  string
	LogIgnoreList         string
	OpenAIKey             string
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

var config GeneralConfig

var log = slog.Scope("MCP2MQTT")

func GetCacheFileName() string {
	return os.Getenv("TW_CACHE_PREFIX") + "cacheclips.bin"
}

func GetConfig() GeneralConfig {
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
	cfg := os.Getenv("TW_CONFIG_PREFIX") + configFile
	log.Info("Loading config %s", cfg)
	if !osutil.Exists(cfg) {
		log.Error("Config file %s does not exists.", cfg)
		os.Exit(1)
	}

	_, err := toml.DecodeFile(cfg, &config)
	if err != nil {
		log.Error("Error decoding file %s: %s", cfg, err)
		os.Exit(1)
	}
}

func SaveConfig() {
	cfg := os.Getenv("TW_CONFIG_PREFIX") + configFile
	log.Info("Saving config %s", cfg)
	f, err := os.Create(cfg)
	if err != nil {
		log.Fatal("Error opening %s: %s", cfg, err)
	}
	e := toml.NewEncoder(f)
	err = e.Encode(&config)
	if err != nil {
		log.Fatal("Error saving data to %s: %s", cfg, err)
	}
}
