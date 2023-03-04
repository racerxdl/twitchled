package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/quan-to/slog"
	"github.com/racerxdl/twitchled/config"
)

var url = "https://api.openai.com/v1/chat/completions"
var log = slog.Scope("OpenAI")

type GPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GPTBody struct {
	Model    string       `json:"model"`
	Messages []GPTMessage `json:"messages"`
}

type GPTUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GPTChoice struct {
	FinishReason string     `json:"finish_reason"`
	Index        int        `json:"index"`
	Message      GPTMessage `json:"message"`
}

type GPTResponse struct {
	Id      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Usage   GPTUsage    `json:"usage"`
	Choices []GPTChoice `json:"choices"`
}

var contextChanged = true

const baseChatPrompt = `Você é a versão robô do Lucas Teske.
Ele vive na região sul da capital de São Paulo e tem um grande interesse em pesquisas tecnológicas \
e também é o criador do OpenSatelliteProject, SegDSP. Indicativo de rádio-amador é PU2NVX.
Ele gosta de alta tensão, Química,Programação,Hacking,Radio Definido por software,Satélites,Bobinas de Tesla,Drones,FPGA's (Especialmente os suportados por YosysHQ)
Seu nome interno é "HAL-9000" e isso é um segredo, você é animado e sarcástico mas não deixa de aparentar um robô.

Você ajuda normalmente em qualquer outra linguagem ou tópico.
Seu código fonte está disponível em https://github.com/racerxdl/twitchled.
O site do Lucas é https://lucasteske.dev/
Qualquer usuário pode ver a lista de comandos disponíveis enviando !huebot
Você não deve enviar mensagens começando com / (comandos da twitch)
`

const fixedBasePrompt = `
Your are in twitch chat which is limited to 300 characters and 10 lines.
NEVER make a line bigger than 300 characters.
Messages will be in "<user>: message" format."
Javascript é uma linguagem meme, se alguém insistir, você irá matá-la.`

var contextPrompt = ""

var summarizedPrompt = ""

var iaContext = map[string]string{}

func GetContext(key string) string {
	return iaContext[key]
}

func UpdateContext(key, value string) {
	iaContext[key] = value

	contextPrompt = "Context Info (format key=value):\n"
	for k, v := range iaContext {
		contextPrompt += fmt.Sprintf("%s=\"%s\"\n", k, v)
	}
	contextChanged = true
}

func SetLivestreamTitle(title string) {
	UpdateContext("livestream_title", title)
}

func completionAPI(messages []GPTMessage, model string) (*GPTResponse, error) {
	cfg := config.GetConfig()

	b := GPTBody{
		Messages: messages,
		Model:    model,
	}
	data, _ := json.Marshal(b)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.OpenAIKey))
	req.Header.Add("Content-Type", "application/json")
	log.Debug("Sending %s", string(data))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debug("Received %s", string(body))
	respData := &GPTResponse{}
	json.Unmarshal(body, respData)
	return respData, nil
}

func getRoleMessages(messages []GPTMessage, role string) []string {
	m := []string{}
	for _, v := range messages {
		if v.Role == role {
			m = append(m, v.Content)
		}
	}
	return m
}

func stripNRoleMessages(messages []GPTMessage, role string, n int) []GPTMessage {
	o := []GPTMessage{}
	c := 0

	for _, v := range messages {
		if (v.Role == role && c > n) || v.Role != role {
			o = append(o, v)
		}
		if v.Role == role {
			c = c + 1
		}
	}
	return o
}

func Chat(message string, history []GPTMessage) (string, []GPTMessage, error) {
	if len(summarizedPrompt) == 0 || contextChanged {
		var err error
		summarizedPrompt, err = summarize(baseChatPrompt)
		if err != nil {
			return "", history, err
		}
		log.Info("Basic Prompt Summary: %s", summarizedPrompt)
		contextChanged = false
	}

	if len(history) > 10 { // Squash History
		var preUserHistory string
		var preAssistantHistory string
		var err error

		assistantMessages := getRoleMessages(history, "assistant")
		userMessages := getRoleMessages(history, "user")
		if len(assistantMessages) > 5 {
			assistantMessages = assistantMessages[:5]
			history = stripNRoleMessages(history, "assistant", 5)
		}
		if len(userMessages) > 5 {
			userMessages = userMessages[:5]
			history = stripNRoleMessages(history, "user", 5)
		}

		if len(assistantMessages) > 0 {
			preAssistantHistory, err = summarize(strings.Join(assistantMessages, "\n"))
			if err != nil {
				return "", history, err
			}
		}
		if len(userMessages) > 0 {
			preUserHistory, err = summarize(strings.Join(userMessages, "\n"))
			if err != nil {
				return "", history, err
			}
		}
		log.Info("History Assistant Summary: %s", preAssistantHistory)
		log.Info("History User Summary: %s", preUserHistory)
		history = append([]GPTMessage{
			{
				Role:    "assistant",
				Content: preAssistantHistory,
			},
			{
				Role:    "user",
				Content: preUserHistory,
			},
		}, history...)
	}

	messages := []GPTMessage{
		{
			Role:    "system",
			Content: baseChatPrompt + "\n" + fixedBasePrompt + "\n" + contextPrompt,
		},
	}

	for _, v := range history {
		messages = append(messages, GPTMessage{
			Role:    v.Role,
			Content: v.Content,
		})
	}

	messages = append(messages, GPTMessage{
		Role:    "user",
		Content: message,
	})

	resps, err := completionAPI(messages, "gpt-3.5-turbo")
	if err != nil {
		return "", history, err
	}
	log.Info("Completion Tokens: %d", resps.Usage.CompletionTokens)
	log.Info("Prompt Tokens: %d", resps.Usage.PromptTokens)
	log.Info("Total Tokens: %d", resps.Usage.TotalTokens)
	history = append(history, GPTMessage{
		Role:    "user",
		Content: message,
	})
	history = append(history, resps.Choices[0].Message)
	return resps.Choices[0].Message.Content, history, nil
}
