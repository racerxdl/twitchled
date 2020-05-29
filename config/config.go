package config

import (
	"github.com/BurntSushi/toml"
	"github.com/mewkiz/pkg/osutil"
	"github.com/quan-to/slog"
	"os"
)

type MQTTConfig struct {
	Host       string
	User       string
	Pass       string
	DeviceName string
}

const configFile = "twitchled.toml"

var config MQTTConfig

var log = slog.Scope("MCP2MQTT")

func GetConfig() MQTTConfig {
	return config
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
