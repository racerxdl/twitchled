package twitch

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"github.com/racerxdl/twitchled/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

var appToken *oauth2.Token

func LoadClientToken() {
	b64data := config.GetConfig().TwitchAppTokenData
	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		log.Error("No app token data on disk or invalid")
		return
	}
	d := gob.NewDecoder(bytes.NewBuffer(data))

	err = d.Decode(&appToken)
	if err != nil {
		log.Error("No app token data on disk or invalid: %s", err)
		return
	}
}

func GetAppToken() string {
	var err error
	if appToken == nil {
		LoadClientToken()
	}

	if appToken.Valid() {
		return appToken.AccessToken
	}

	c := config.GetConfig()
	cfg := &clientcredentials.Config{
		ClientID:     c.TwitchOAuthClient,
		ClientSecret: c.TwitchOAuthSecret,
		TokenURL:     twitch.Endpoint.TokenURL,
	}

	appToken, err = cfg.Token(context.Background())
	if err != nil {
		log.Error("Error getting token: %s", err)
		return ""
	}

	buff := bytes.NewBuffer(nil)
	e := gob.NewEncoder(buff)
	e.Encode(&appToken)

	config.SetTwitchAppTokenData(buff.Bytes())

	return appToken.AccessToken
}
