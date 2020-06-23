package main

import (
	"fmt"
	"github.com/racerxdl/twitchled/twitch"
	"github.com/racerxdl/twitchled/wimatrix"
	"golang.org/x/image/colornames"
	"image/color"
	"strconv"
	"strings"
)

const (
	cmdHelp     = "!huebot"
	cmdHelpCmd  = "!hue"
	cmdColor    = "!color"
	cmdBgColor  = "!bgcolor"
	cmdBright   = "!bright"
	cmdBgBright = "!bgbright"
	cmdSource   = "!source"
	cmdPanel    = "!painel"
	cmdSpeed    = "!speed"
	cmdLight    = "!light"
)

func isCommand(cmd, msg string) bool {
	return len(msg) >= len(cmd) && msg[:len(cmd)] == cmd
}

func ParseChat(chat *twitch.Chat, event *twitch.MessageEventData) {
	log.Info("User %s: %s", event.Username, event.Message)
	if isCommand(cmdHelp, event.Message) {
		chat.SendMessage(fmt.Sprintf("Olá @%s! Quer fazer uns HUEHUE? - Use !hue COMANDO para ajuda de um comando. Comandos: !color !bgcolor !bright !bgbright !source", event.Username))
		return
	}

	if isCommand(cmdHelpCmd, event.Message) {
		cmdName := strings.Trim(event.Message[len(cmdHelpCmd):], " !")
		CmdHelp(chat, event.Username, cmdName)
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
		chat.SendMessage(fmt.Sprintf("Olá @%s! Meu código fonte está no Github! https://github.com/racerxdl/twitchled - E o código fonte do painel de LED também: https://github.com/racerxdl/wimatrix", event.Username))
		return
	}

	//if isCommand(cmdPanel, event.Message) {
	//	CmdMessage(event.Username, event.Message[len(cmdPanel):])
	//	return
	//}

	if isCommand(cmdSpeed, event.Message) {
		CmdSpeed(event.Message[len(cmdSpeed):])
		return
	}

	//if isCommand(cmdLight, event.Message) {
	//	CmdLight()
	//	return
	//}
}

func CmdHelp(chat *twitch.Chat, username, cmdName string) {
	switch cmdName {
	case "color":
		chat.SendMessage(fmt.Sprintf("@%s, o comando color troca a cor do texto! Você pode dar o nome da cor ou em hexa. Por exemplo !color red ou !color #FF0000", username))
	case "bgcolor":
		chat.SendMessage(fmt.Sprintf("@%s, o comando color troca a cor do fundo! Você pode dar o nome da cor ou em hexa. Por exemplo !bgcolor red ou !bgcolor #FF0000", username))
	case "bright":
		chat.SendMessage(fmt.Sprintf("@%s, o comando color troca o brilho do texto! O valor mínimo é 0 e máximo é 1. Você pode usar !bright 1", username))
	case "bgbright":
		chat.SendMessage(fmt.Sprintf("@%s, o comando color troca o brilho do fundo! O valor mínimo é 0 e máximo é 1. Você pode usar !bright 1", username))
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
