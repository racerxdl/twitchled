package twitch

import (
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/irc.v3"
	"strings"
	"time"
)

const (
	channelBufferSize = 16
	ChatTLS           = "irc.chat.twitch.tv:6697"
)

type Chat struct {
	id          string
	channelName string
	conn        *tls.Conn
	ircClient   *irc.Client

	Events chan ChatEvent
}

func MakeChat(botName, channelName, chatToken string) (*Chat, error) {
	c := &Chat{
		id:          uuid.New().String(),
		channelName: fmt.Sprintf("#%s", channelName),
		Events:      make(chan ChatEvent, channelBufferSize),
	}

	log.Info("Connecting to %s", ChatTLS)
	conn, err := tls.Dial("tcp", ChatTLS, &tls.Config{})
	if err != nil {
		return nil, err
	}

	c.conn = conn

	log.Info("Connected to chat. Starting IRC")
	c.ircClient = irc.NewClient(conn, irc.ClientConfig{
		Nick:    botName,
		User:    botName,
		Name:    botName,
		Pass:    "oauth:" + chatToken,
		Handler: irc.HandlerFunc(c.ircHandler),
	})

	go c.runIRC()

	now := time.Now()

	err = fmt.Errorf("timeout waiting login")

	for time.Since(now) < time.Second*5 {
		doBreak := false
		e := <-c.Events
		switch e.GetType() {
		case EventError:
			err = e.GetData().(*ErrorEventData).RawError()
			doBreak = true
		case EventLoginError:
			err = fmt.Errorf(e.GetData().(*LoginEventData).Message)
			doBreak = true
		case EventLoginSuccess:
			err = nil
			doBreak = true
		}

		if doBreak {
			break
		}
	}

	if err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Chat) runIRC() {
	log.Debug("Running IRC Client")
	err := c.ircClient.Run()
	if err != nil {
		log.Error("Error in IRC Client: ", err)
		c.Events <- MakeErrorEvent(err)
	}
}

func (c *Chat) SendMessage(msg string) error {
	return c.ircClient.WriteMessage(&irc.Message{
		Params:  []string{c.channelName, msg},
		Command: "PRIVMSG",
	})
}

func (c *Chat) SendRawMessage(msg string) error {
	return c.ircClient.Write(msg)
}

func (c *Chat) ircHandler(ircClient *irc.Client, m *irc.Message) {
	switch m.Command {
	case "001": // Welcome
		log.Debug("Joining channel %s", c.channelName)
		c.SendRawMessage(fmt.Sprintf("JOIN %s", c.channelName))
		c.Events <- MakeLoginEvent(true, m.Params[1])
	case "002":
	case "003":
	case "004":
	case "005":
	case "250":
	case "251": // User Report
		log.Info(m.Params[1])
	case "252":
	case "253":
	case "254":
	case "255":
	case "265":
	case "266":
	case "353": // Name List
	case "366": // End of Name List
	case "375": // MOTD Start
		//log.Debug("MOTD: %s", m.Params[1])
	case "372": // MOTD Body
		//log.Debug("MOTD: %s", m.Params[1])
	case "376": // MOTD End
	case "PRIVMSG":
		if ircClient.FromChannel(m) && len(m.Params) >= 2 {
			// channel := m.Params[0]
			message := m.Params[1]
			from := m.User
			picture := ""
			pic, err := GetProfilePic(from)
			if err == nil {
				picture = pic
			}
			c.Events <- MakeMessageEventData(SourceTwitch, from, message, picture, m)
		}
	case "NOTICE":
		//log.Info("[%s] %s {{%+v}}", m.Command, m.String(), m.Params)
		if strings.Contains(m.Params[1], "Login authentication failed") {
			// Login failed
			c.Events <- MakeLoginEvent(false, m.Params[1])
		}
	case "JOIN":
		//log.Debug("JOIN: %s joins %s", m.User, m.Params[0])
	case "PING":
	default:
		//log.Debug("[%s] %s {{%+v}}", m.Command, m.String(), m.Params)
	}
}

func (c *Chat) EventChannel() chan ChatEvent {
	return c.Events
}

func (c *Chat) Id() string {
	return c.id
}
