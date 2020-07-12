package websub

import (
	"crypto/hmac"
	"crypto/sha256"
	"regexp"
	"strings"
)

var (
	extractUrl = regexp.MustCompile(`.*<(.*)>.*`)
)

func getChannelId(link string) string {
	urls := strings.Split(link, ",")
	theUrl := ""

	for _, v := range urls {
		s := extractUrl.FindAllStringSubmatch(v, -1)
		if len(s) > 0 {
			t := s[0]
			if len(t) > 1 {
				if !strings.Contains(t[1], "webhooks/hub") {
					theUrl = t[1]
					break
				}
			}
		}
	}

	ft := strings.Replace(followsTopic, "%s", "", -1)
	ss := strings.Replace(streamStatusTopic, "%s", "", -1)
	if strings.Contains(theUrl, ft) {
		return theUrl[len(ft):]
	}

	if strings.Contains(theUrl, ss) {
		return theUrl[len(ss):]
	}

	return ""
}

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha256.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}
