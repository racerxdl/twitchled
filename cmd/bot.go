package main

import (
	"fmt"
	"github.com/racerxdl/twitchled/discord"
	"github.com/racerxdl/twitchled/twitch"
	"github.com/racerxdl/twitchled/twitch/twitchdata"
	"github.com/racerxdl/twitchled/wimatrix"
	"golang.org/x/image/colornames"
	"image/color"
	"strconv"
	"strings"
	"time"
)

const (
	cmdHelp           = "!huebot"
	cmdHelpCmd        = "!hue"
	cmdHelpEnglishCmd = "!huenglish"
	cmdColor          = "!color"
	cmdBgColor        = "!bgcolor"
	cmdBright         = "!bright"
	cmdBgBright       = "!bgbright"
	cmdSource         = "!source"
	cmdPanel          = "!painel"
	cmdSpeed          = "!speed"
	cmdLight          = "!light"
	cmdCommands       = "!comandos"
	cmdMode           = "!panelmode"
	textBoaNoite      = "boa noite"
	textBomDia        = "boa dia"
	textGoodNight     = "good night"
	textGoodMorning   = "good morning"
	fakeSub           = "fake sub"
	fakeBits          = "fake bits"
	fakeFollow        = "fake follow"
)

var allcmds = []string{
	cmdHelp, cmdHelpCmd, cmdColor, cmdBgColor, cmdBright, cmdBgBright, cmdSource, cmdPanel, cmdSpeed, cmdLight,
}

var subOnlyCmds = []string{
	cmdLight,
	cmdPanel,
}

func isCommand(cmd, msg string) bool {
	return len(msg) >= len(cmd) && msg[:len(cmd)] == cmd
}

func ParseChat(chat *twitch.Chat, event *twitch.MessageEventData) {
	log.Info("User %s: %s", event.Username, event.Message)
	userPrefix := ""

	discord.Log(event.Username, event.Picture, event.Message)

	if event.IsSubscriber() {
		userPrefix = "Doctor"
	}

	if event.IsModerator() {
		userPrefix = "Moderator"
	}

	if isCommand(cmdHelp, event.Message) || isCommand(cmdCommands, event.Message) {
		_ = chat.SendMessage(fmt.Sprintf("Olá %s @%s! Quer fazer uns HUEHUE? - Use !hue COMANDO para ajuda de um comando. Comandos: %s", userPrefix, event.Username, strings.Join(allcmds, " ")))
		_ = chat.SendMessage(fmt.Sprintf("Hi %s @%s! Want to do some HUEHUE? - Use !huenglish COMMAND for help of a command. Commands: %s", userPrefix, event.Username, strings.Join(allcmds, " ")))
		return
	}

	if isCommand(cmdHelpEnglishCmd, event.Message) {
		cmdName := strings.Trim(event.Message[len(cmdHelpEnglishCmd):], " !")
		CmdHelpEnglish(chat, userPrefix, event.Username, cmdName)
		return
	}

	if isCommand(cmdHelpCmd, event.Message) {
		cmdName := strings.Trim(event.Message[len(cmdHelpCmd):], " !")
		CmdHelp(chat, userPrefix, event.Username, cmdName)
		return
	}

	if isCommand(cmdColor, event.Message) {
		CmdColor(event.Message[len(cmdColor):])
		return
	}

	if isCommand(cmdBgColor, event.Message) {
		CmdBGColor(event.Message[len(cmdBgColor):])
		return
	}

	if isCommand(cmdBgBright, event.Message) {
		CmdBGBright(event.Message[len(cmdBgBright):])
		return
	}

	if isCommand(cmdBright, event.Message) {
		CmdBright(event.Message[len(cmdBright):])
		return
	}

	if isCommand(cmdSource, event.Message) {
		_ = chat.SendMessage(fmt.Sprintf("Olá %s @%s! Meu código fonte está no Github! https://github.com/racerxdl/twitchled - E o código fonte do painel de LED também: https://github.com/racerxdl/wimatrix", userPrefix, event.Username))
		_ = chat.SendMessage(fmt.Sprintf("Hello %s @%s! My source code is in Github! https://github.com/racerxdl/twitchled - And the led panel as well: https://github.com/racerxdl/wimatrix", userPrefix, event.Username))
		return
	}

	if isCommand(cmdSpeed, event.Message) {
		CmdSpeed(event.Message[len(cmdSpeed):])
		return
	}

	if strings.Contains(strings.ToLower(event.Message), textBoaNoite) {
		_ = chat.SendMessage(fmt.Sprintf("Boa noite %s @%s!", userPrefix, event.Username))
		return
	}

	if strings.Contains(strings.ToLower(event.Message), textBomDia) {
		_ = chat.SendMessage(fmt.Sprintf("Boa dia %s @%s!", userPrefix, event.Username))
		return
	}

	if strings.Contains(strings.ToLower(event.Message), textGoodMorning) {
		_ = chat.SendMessage(fmt.Sprintf("Good morning %s @%s!", userPrefix, event.Username))
		return
	}

	if strings.Contains(strings.ToLower(event.Message), textGoodNight) {
		_ = chat.SendMessage(fmt.Sprintf("Good night %s @%s!", userPrefix, event.Username))
		return
	}

	if event.IsSubscriber() || event.IsModerator() {
		// Subscriber only events
		if isCommand(cmdPanel, event.Message) {
			CmdMessage(event.Username, event.Message[len(cmdPanel):])
			return
		}

		if isCommand(cmdLight, event.Message) {
			CmdLight()
			return
		}
	} else {
		for _, v := range subOnlyCmds {
			if isCommand(v, event.Message) {
				_ = chat.SendMessage(fmt.Sprintf("Desculpe %s %s, %s é apenas para subs.", userPrefix, event.Username, v))
				_ = chat.SendMessage(fmt.Sprintf("Sorry %s %s, %s is for subscribers only.", userPrefix, event.Username, v))
				return
			}
		}
	}

	// Owner Only
	if strings.ToLower(event.Username) == "racerxdl" || event.IsModerator() {
		// OWNER

		if isCommand(fakeSub, event.Message) {
			_ = chat.SendMessage("OK my king. A new fake subscription is coming")
			e := twitch.MakeSubscribeEventData("FAKE", twitchdata.ChannelSubscribeMessageData{
				UserName:    "JohnCena",
				DisplayName: "JohnCena",
				SubMessage: twitchdata.ChannelSubscriberMessage{
					Message: "MY NAME IS JOHN CENA!!!",
				},
				StreakMonths:     1000,
				CumulativeMonths: 1000,
				ChatMessage:      "MY NAME IS JOHN CENA!!!",
			})
			OnSub(chat, e.(*twitch.SubscribeEventData))
			return
		}

		if isCommand(fakeBits, event.Message) {
			_ = chat.SendMessage("OK my king. A new fake bits")
			e := twitch.MakeBitsV2EventData("FAKE", twitchdata.BitEventsV2{
				IsAnonymous: false,
				Data: twitchdata.BitEventsData{
					UserName:      "JohnCena",
					BitsUsed:      1000000,
					TotalBitsUsed: 1000000,
					ChatMessage:   "MY NAME IS JOHN CENA!!!",
				},
			})
			OnBits(chat, e.(*twitch.BitsV2EventData))
			return
		}

		if isCommand(fakeFollow, event.Message) {
			_ = chat.SendMessage("OK my king. A new fake follow")
			e := twitch.MakeFollowEventData("FAKE", "JohnScena", "1000000")
			OnFollow(chat, e.(*twitch.FollowEventData))
			return
		}

		if isCommand(cmdMode, event.Message) {
			data := event.Message[len(cmdMode):]
			data = strings.Trim(data, " \r\n")

			printValidModes := func() {
				validModesString := "Valid modes are: "
				for _, v := range wimatrix.Modes {
					validModesString += fmt.Sprintf("%d: %s ", int(v), v.String())
				}
				_ = chat.SendMessage(validModesString)
				log.Debug(validModesString)
			}

			v, err := strconv.Atoi(data)
			if err != nil {
				_ = chat.SendMessage(fmt.Sprintf("Invalid mode %q: %s", data, err))
				printValidModes()
				return
			}
			ok := false
			for _, mode := range wimatrix.Modes {
				if v == int(mode) {
					ok = true
					break
				}
			}

			if !ok {
				_ = chat.SendMessage(fmt.Sprintf("Invalid mode %q.", data))
				printValidModes()
				return
			}

			ev.Publish(wimatrix.EvNewMode, wimatrix.Mode(v))
			_ = chat.SendMessage(fmt.Sprintf("Mode set to %d: %s", v, wimatrix.Mode(v).String()))
		}
	}

	if strings.Index(event.Message, "javascripto") >= 0 {
		if time.Since(lastJavascriptoPlay) < time.Second*30 {
			return
		}

		lastJavascriptoPlay = time.Now()
		log.Info("MATA O JAVASCRIPTO!!!")
		_ = chat.SendMessage("/me EU AVO MATA O JAVASCRIPTOOOO!!!!!")
		go func() {
			err := PlayJavascripto()
			if err != nil {
				log.Error("Error playing audio %q", err)
			}
		}()
	}
}

func CmdHelp(chat *twitch.Chat, userPrefix, username, cmdName string) {
	if len(cmdName) > 0 && cmdName[0] == '!' {
		cmdName = cmdName[1:]
	}

	switch cmdName {
	case "huebot":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, eu sou o bot do @RacerXDL!", userPrefix, username))
	case "hue":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, o que deseja saber?", userPrefix, username))
	case "color":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, o comando color troca a cor do texto! Você pode dar o nome da cor ou em hexa. Por exemplo !color red ou !color #FF0000", userPrefix, username))
	case "bgcolor":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, o comando bgcolor troca a cor do fundo! Você pode dar o nome da cor ou em hexa. Por exemplo !bgcolor red ou !bgcolor #FF0000", userPrefix, username))
	case "bright":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, o comando bright troca o brilho do texto! O valor mínimo é 0 e máximo é 1. Você pode usar !bright 1", userPrefix, username))
	case "bgbright":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, o comando bgbright troca o brilho do fundo! O valor mínimo é 0 e máximo é 1. Você pode usar !bgbright 1", userPrefix, username))
	case "source":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, o comando source mostra o meu código fonte e o do painel de led!", userPrefix, username))
	case "painel":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, se você for subscriber, o comando painel envia uma mensagem no painel de led! Por exemplo: !painel HUEBOT é muito legal", userPrefix, username))
	case "speed":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, se você for subscriber, o comando speed muda a velocidade da mensagem no painel! Por exemplo: !speed 60", userPrefix, username))
	case "light":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, se você for subscriber, o comando light aperta o interruptor da luz do quarto do @RacerXDL!", userPrefix, username))
	default:
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, desculpa, mas eu não conheço o comando %q :(", userPrefix, username, cmdName))
	}
}

func CmdHelpEnglish(chat *twitch.Chat, userPrefix, username, cmdName string) {
	if len(cmdName) > 0 && cmdName[0] == '!' {
		cmdName = cmdName[1:]
	}

	switch cmdName {
	case "huebot":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, I'm @RacerXDL bot!", userPrefix, username))
	case "hue":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, what do you want to know?", userPrefix, username))
	case "color":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, the command changes the text color! You can give the name of the color or hex value. For example !color red or !color #FF0000", userPrefix, username))
	case "bgcolor":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, the command changes the background color! You can give the name of the color or hex value. For example !bgcolor red or !bgcolor #FF0000", userPrefix, username))
	case "bright":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, the command changes text brightness! The minimum value is 0 and maximum is 1. You can use !bright 1", userPrefix, username))
	case "bgbright":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, the command changes background brightness! The minimum value is 0 and maximum is 1. You can use !bgbright 1", userPrefix, username))
	case "source":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, the command soruce shows mine and led panel source code!", userPrefix, username))
	case "painel":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, if you're a subscriber, sends a text message to the led panel! For example !painel HUEBOT is so cool", userPrefix, username))
	case "speed":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, if you're a subscriber, changes the scrolling speed of the text in the panel! For example !speed 60", userPrefix, username))
	case "light":
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, if you're a subscriber, the command light toggles the room light of @RacerXDL!", userPrefix, username))
	default:
		_ = chat.SendMessage(fmt.Sprintf("%s @%s, sorry, but I don't know the command %q :(", userPrefix, username, cmdName))
	}
}

func parseColorFromMessage(msg string) (color.Color, error) {
	if msg[0] == '#' { // Hex color
		msg = msg[1:]
		ci, err := strconv.ParseInt(msg, 16, 32)
		if err != nil {
			return color.Black, fmt.Errorf("invalid color")
		}

		c := color.RGBA{
			R: uint8((ci & 0xFF0000) >> 16),
			G: uint8((ci & 0x00FF00) >> 8),
			B: uint8((ci & 0x0000FF) >> 0),
			A: 255,
		}

		return c, nil

	} else { // Name color
		if c, ok := colornames.Map[msg]; ok {
			return c, nil
		}
	}

	return color.Black, fmt.Errorf("invalid color")
}

func CmdColor(msg string) {
	msg = strings.Trim(msg, " !")
	if len(msg) < 2 {
		return
	}

	c, err := parseColorFromMessage(msg)
	if err != nil {
		return
	}
	ev.Publish(wimatrix.EvSetTextColor, c)
}

func CmdBGColor(msg string) {
	msg = strings.Trim(msg, " !")
	if len(msg) < 2 {
		return
	}

	c, err := parseColorFromMessage(msg)
	if err != nil {
		return
	}
	ev.Publish(wimatrix.EvSetBgColor, c)
}

func CmdBright(msg string) {
	msg = strings.Trim(msg, " !")
	if len(msg) < 1 {
		return
	}

	bright, err := strconv.ParseFloat(msg, 32)
	if err != nil {
		return
	}

	ev.Publish(wimatrix.EvSetTextBrightness, float32(bright))
}

func CmdBGBright(msg string) {
	msg = strings.Trim(msg, " !")
	if len(msg) < 1 {
		return
	}

	bright, err := strconv.ParseFloat(msg, 32)
	if err != nil {
		return
	}

	ev.Publish(wimatrix.EvSetBgBrightness, float32(bright))
}

func CmdMessage(user, msg string) {
	msg = strings.Trim(msg, " !")
	if len(msg) < 1 {
		return
	}

	ev.Publish(wimatrix.EvNewMsg, fmt.Sprintf("%s by %s", msg, user))
}

func CmdSpeed(msg string) {
	msg = strings.Trim(msg, " !")
	if len(msg) < 1 {
		return
	}

	v, err := strconv.Atoi(msg)

	if err != nil {
		return
	}

	ev.Publish(wimatrix.EvSetSpeed, v)
}

func CmdLight() {
	ev.Publish(wimatrix.EvSetLight)
}
