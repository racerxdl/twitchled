package twitch

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pkg/browser"
	"github.com/racerxdl/twitchled/config"
	"github.com/racerxdl/twitchled/discord"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

const (
	stateCallbackKey = "oauth-state-callback"
	oauthSessionName = "oauth-session"
	oauthTokenKey    = "oauth-token"

	HelixAPI = "https://api.twitch.tv/helix"
)

var (
	l            net.Listener
	oauth2Config *oauth2.Config
	cookieSecret = []byte("ABCDE")
	cookieStore  = sessions.NewCookieStore(cookieSecret)
	token        *oauth2.Token
)

func init() {
	reset()
}

func reset() {
	cookieSecret := make([]byte, 32)
	_, _ = rand.Read(cookieSecret)
	cookieStore = sessions.NewCookieStore(cookieSecret)
}

// HandleRoot is a Handler that shows a login button. In production, if the frontend is served / generated
// by Go, it should use html/template to prevent XSS attacks.
func HandleRoot(w http.ResponseWriter, r *http.Request) (err error) {
	reset()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html><body><a href="/login">Login using Twitch</a></body></html>`))

	return
}

// HandleLogin is a Handler that redirects the user to Twitch for login, and provides the 'state'
// parameter which protects against login CSRF.
func HandleLogin(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := cookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Error("corrupted session %s -- generated new", err)
		err = nil
	}

	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return AnnotateError(err, "Couldn't generate a session!", http.StatusInternalServerError)
	}

	state := hex.EncodeToString(tokenBytes[:])

	session.AddFlash(state, stateCallbackKey)

	if err = session.Save(r, w); err != nil {
		return
	}

	http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)

	return
}

// HandleOauth2Callback is a Handler for oauth's 'redirect_uri' endpoint;
// it validates the state token and retrieves an OAuth token from the request parameters.
func HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := cookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Error("corrupted session %s -- generated new", err)
		err = nil
	}

	// ensure we flush the csrf challenge even if the request is ultimately unsuccessful
	defer func() {
		if err := session.Save(r, w); err != nil {
			log.Error("error saving session: %s", err)
		}
	}()

	switch stateChallenge, state := session.Flashes(stateCallbackKey), r.FormValue("state"); {
	case state == "", len(stateChallenge) < 1:
		err = errors.New("missing state challenge")
	case state != stateChallenge[0]:
		err = fmt.Errorf("invalid oauth state, expected %q, got %q", state, stateChallenge[0])
	}

	if err != nil {
		return AnnotateError(
			err,
			"Couldn't verify your confirmation, please try again.",
			http.StatusBadRequest,
		)
	}

	token, err = oauth2Config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		return
	}

	// add the oauth token to session
	session.Values[oauthTokenKey] = token

	if token.Valid() {
		SaveToken()
	}

	http.Redirect(w, r, "/done", http.StatusTemporaryRedirect)

	l.Close()

	return
}

func HandleDone(w http.ResponseWriter, r *http.Request) (err error) {
	w.WriteHeader(200)
	w.Write([]byte(`
	<script type="text/javascript">
	// ==UserScript==
	// @name        window.close demo
	// @include     http://localhost:7001/*
	// @grant       GM_addStyle
	// ==/UserScript==
	window.close();
	if (top) {
		top.close();
	}
	</script>
	You can now close this window
	`))

	l.Close()

	return nil
}

// HumanReadableError represents error information
// that can be fed back to a human user.
//
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type HumanReadableError interface {
	HumanError() string
	HTTPCode() int
}

// HumanReadableWrapper implements HumanReadableError
type HumanReadableWrapper struct {
	ToHuman string
	Code    int
	error
}

func (h HumanReadableWrapper) HumanError() string { return h.ToHuman }
func (h HumanReadableWrapper) HTTPCode() int      { return h.Code }

// AnnotateError wraps an error with a message that is intended for a human end-user to read,
// plus an associated HTTP error code.
func AnnotateError(err error, annotation string, code int) error {
	if err == nil {
		return nil
	}
	return HumanReadableWrapper{ToHuman: annotation, error: err}
}

type Handler func(http.ResponseWriter, *http.Request) error

func SaveToken() {
	buff := bytes.NewBuffer(nil)
	e := gob.NewEncoder(buff)
	e.Encode(&token)

	config.SetTwitchToken(buff.Bytes())
}

func LoadToken() {
	b64data := config.GetConfig().TwitchTokenData
	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		log.Error("No token data on disk or invalid")
		return
	}
	d := gob.NewDecoder(bytes.NewBuffer(data))

	err = d.Decode(&token)
	if err != nil {
		log.Error("No token data on disk or invalid: %s", err)
		return
	}
}

func RefreshToken() {
	if token != nil && token.RefreshToken != "" {
		_, err := GetChannelId()
		if err == nil {
			token.Expiry = time.Now().Add(time.Hour) // Token is valid, force to check again in a hour
			return
		}
		data := url.Values{}
		data.Add("client_id", config.GetConfig().TwitchOAuthClient)
		data.Add("client_secret", config.GetConfig().TwitchOAuthSecret)
		data.Add("grant_type", "refresh_token")
		data.Add("refresh_token", token.RefreshToken)

		client := &http.Client{}
		r, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", strings.NewReader(data.Encode())) // URL-encoded payload
		if err != nil {
			log.Fatal(err)
		}
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		res, err := client.Do(r)
		if err != nil {
			discord.Log("ERROR", "", fmt.Sprintf("cannot renew token: %q", err))
			log.Fatal(err)
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			discord.Log("ERROR", "", fmt.Sprintf("cannot renew token: %q", err))
			log.Fatal(err)
		}
		var d map[string]interface{}
		err = json.Unmarshal(body, &d)
		if err != nil {
			discord.Log("ERROR", "", fmt.Sprintf("cannot renew token: %q", err))
			log.Fatal(err)
		}
		accessToken := d["access_token"].(string)
		refreshToken := d["refresh_token"].(string)

		token.AccessToken = accessToken
		token.RefreshToken = refreshToken
		token.Expiry = time.Now().Add(time.Second * 3600)
		SaveToken()
	}
}

func GetAccessToken() (*oauth2.Token, error) {
	if token == nil {
		LoadToken()
	}

	if token.Valid() {
		return token, nil
	}

	if !token.Valid() {
		log.Info("Token not valid. Trying to refresh token...")
		RefreshToken()
		if token.Valid() {
			return token, nil
		}
	}

	reset()

	var err error
	gob.Register(&oauth2.Token{})

	oauth2Config = &oauth2.Config{
		ClientID:     config.GetConfig().TwitchOAuthClient,
		ClientSecret: config.GetConfig().TwitchOAuthSecret,
		Scopes: []string{
			"moderation:read",
			"moderator:read:followers",
			"moderator:read:guest_star",
			"moderator:read:shield_mode",
			"moderator:read:shoutouts",

			"channel_read",
			"channel_check_subscription",
			"channel:read:subscriptions",
			"channel:read:redemptions",

			"channel_commercial",
			"channel_feed_read",
			"channel_feed_edit",
			"channel_subscriptions",
			"channel:moderate",
			"channel:read:guest_star",
			"channel:read:polls",
			"channel:read:predictions",
			"channel:read:hype_train",
			"channel:read:charity",
			"channel:read:goals",

			"bits:read",
			"chat:read",
			"chat:edit",
		},
		Endpoint:    twitch.Endpoint,
		RedirectURL: "http://localhost:7001/redirect",
	}

	var middleware = func(h Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			// parse POST body, limit request size
			if err = r.ParseForm(); err != nil {
				return AnnotateError(err, "Something went wrong! Please try again.", http.StatusBadRequest)
			}

			return h(w, r)
		}
	}
	// errorHandling is a middleware that centralises error handling.
	// this prevents a lot of duplication and prevents issues where a missing
	// return causes an error to be printed, but functionality to otherwise continue
	// see https://blog.golang.org/error-handling-and-go
	var errorHandling = func(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := handler(w, r); err != nil {
				var errorString = "Something went wrong! Please try again."
				var errorCode = 500

				if v, ok := err.(HumanReadableError); ok {
					errorString, errorCode = v.HumanError(), v.HTTPCode()
				}

				log.Error("HTTP ERROR: %s", err)
				w.Write([]byte(errorString))
				w.WriteHeader(errorCode)
				return
			}
		})
	}

	var handleFunc = func(path string, handler Handler) {
		http.Handle(path, errorHandling(middleware(handler)))
	}

	handleFunc("/", HandleRoot)
	handleFunc("/login", HandleLogin)
	handleFunc("/redirect", HandleOAuth2Callback)
	handleFunc("/done", HandleDone)

	log.Info("Open up http://localhost:7001 on your browser")

	l, err = net.Listen("tcp", ":7001")
	if err != nil {
		log.Error("Error getting token: %s", err)
		return nil, err
	}

	go browser.OpenURL("http://localhost:7001")

	_ = http.Serve(l, nil)

	if token == nil {
		log.Error("Cannot get token")
		return nil, fmt.Errorf("cannot get token")
	}

	return token, nil
}

func GetChannel(name string) (channelId, channelName string, err error) {
	token, err := GetAccessToken()
	if err != nil {
		return "", "", err
	}

	fullUrl := fmt.Sprintf("https://api.twitch.tv/kraken/users?login=%s", url.QueryEscape(name))

	u, err := url.Parse(fullUrl)

	if err != nil {
		return "", "", err
	}

	req, _ := http.NewRequest("GET", u.String(), nil)

	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", token.AccessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("http error (%d) %s", res.StatusCode, res.Status)
	}

	defer res.Body.Close()

	rawData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	obj := map[string]interface{}{}

	err = json.Unmarshal(rawData, &obj)

	if err != nil {
		return "", "", err
	}

	total, ok := obj["_total"].(float64)

	if !ok {
		return "", "", fmt.Errorf("_total field not found")
	}

	if total == 0 {
		return "", "", fmt.Errorf("channel %s not found", name)
	}

	userI := obj["users"].([]interface{})
	user := userI[0].(map[string]interface{})

	channelId = user["_id"].(string)
	channelName = user["name"].(string)

	return channelId, channelName, nil
}

func GetChannelId() (string, error) {
	if token == nil {
		return "", fmt.Errorf("invalid token")
	}

	fullUrl := "https://api.twitch.tv/helix/users"

	u, err := url.Parse(fullUrl)

	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error (%d) %s", res.StatusCode, res.Status)
	}

	defer res.Body.Close()

	rawData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	obj := map[string]interface{}{}
	err = json.Unmarshal(rawData, &obj)

	if err != nil {
		return "", err
	}

	data := obj["data"].([]interface{})
	data0 := data[0].(map[string]interface{})

	if id, ok := data0["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("cannot find id field")
}

func GetChannelName() (string, error) {
	token, err := GetAccessToken()
	if err != nil {
		return "", err
	}

	fullUrl := "https://api.twitch.tv/helix/users"

	u, err := url.Parse(fullUrl)

	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("GET", u.String(), nil)

	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error (%d) %s", res.StatusCode, res.Status)
	}

	defer res.Body.Close()

	rawData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	obj := map[string]interface{}{}

	err = json.Unmarshal(rawData, &obj)

	if err != nil {
		return "", err
	}

	data := obj["data"].([]interface{})
	data0 := data[0].(map[string]interface{})

	if id, ok := data0["display_name"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("cannot find _id field")
}

func Get(path string) (map[string]interface{}, error) {
	fullUrl := fmt.Sprintf("%s%s", HelixAPI, path)

	u, err := url.Parse(fullUrl)

	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", u.String(), nil)

	req.Header.Add("Client-ID", config.GetConfig().TwitchOAuthClient)
	req.Header.Add("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error (%d) %s", res.StatusCode, res.Status)
	}

	defer res.Body.Close()

	rawData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	obj := map[string]interface{}{}

	err = json.Unmarshal(rawData, &obj)

	return obj, err
}

func GetClips(channelId string, since time.Time) ([]string, error) {
	now := since.Format(time.RFC3339)
	data, err := Get(fmt.Sprintf("/clips?broadcaster_id=%s&started_at=%s", url.PathEscape(channelId), url.PathEscape(now)))
	if err != nil {
		return nil, err
	}

	clips, ok := data["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected data field")
	}

	var urls []string

	for _, v := range clips {
		clip := v.(map[string]interface{})
		clipUrl, ok := clip["url"].(string)
		if !ok {
			continue
		}
		clipTime, ok := clip["created_at"].(string)
		if ok {
			clipTimeParsed, err := time.Parse(time.RFC3339, clipTime)
			if err == nil && time.Since(clipTimeParsed) > time.Minute {
				urls = append(urls, clipUrl)
			}
		}
	}

	return urls, nil
}

func GetProfilePic(channelId string) (string, error) {
	data, err := Get(fmt.Sprintf("/users?login=%s", url.PathEscape(channelId)))

	if err != nil {
		return "", err
	}

	users, ok := data["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("expected data field")
	}

	if len(users) == 0 {
		return "", fmt.Errorf("not found")
	}

	user := users[0].(map[string]interface{})
	logoI := user["profile_image_url"]

	if logoI == nil {
		return "", fmt.Errorf("no logo found")
	}

	logo := logoI.(string)
	return logo, nil
}

//func GetFollowers(channelId string) ([]Follower, error) {
//	data, err := Get(fmt.Sprintf("/channels/%s/follows", channelId))
//
//	if err != nil {
//		return nil, err
//	}
//
//	followers := make([]Follower, 0)
//	follows := data["follows"].([]interface{})
//
//	for _, v := range follows {
//		f := MakeFollowerFromJSON(v.(map[string]interface{}))
//		followers = append(followers, f)
//	}
//
//	return followers, nil
//}
