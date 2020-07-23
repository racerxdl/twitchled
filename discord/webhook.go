package discord

import (
    "bytes"
    "encoding/json"
    "github.com/quan-to/slog"
    "github.com/racerxdl/twitchled/config"
    "io/ioutil"
    "net/http"
)

var log = slog.Scope("Discord")

type allowedMentions struct {
    Parse []string `json:"parse"`
    Users []string `json:"users"`
}

type payload struct {
    Content         string          `json:"content"`
    Username        string          `json:"username"`
    AvatarUrl       string          `json:"avatar_url,omit_empty"`
    AllowedMentions allowedMentions `json:"allowed_mentions"`
}

func send(url, username, avatar, content string) {
    p := payload{
        Content:   content,
        Username:  username,
        AvatarUrl: avatar,
        AllowedMentions: allowedMentions{
            Parse: []string{"users", "roles", "everyone"},
        },
    }

    if config.IsOnIgnoreList(username) {
        return
    }

    jsonStr, _ := json.Marshal(&p)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

    if err != nil {
        log.Error("error creating request: %s", err)
        return
    }

    //req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error("error sending request: %s", err)
        return
    }
    defer resp.Body.Close()

    log.Debug("Discord response status: %d", resp.StatusCode)
    _, _ = ioutil.ReadAll(resp.Body)
}

func Log(username, avatar, message string) {
    c := config.GetConfig()
    if c.DiscordLogOutputUrl == "" {
        return
    }

    send(c.DiscordLogOutputUrl, username, avatar, message)
}

func SendMessage(username, avatar, content string) {
    c := config.GetConfig()
    if c.DiscordBotOutputUrl == "" {
        log.Error("no discord url defined")
        return
    }

    url := c.DiscordBotOutputUrl
    send(url, username, avatar, content)
}
