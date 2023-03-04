package websub

import (
	"crypto/hmac"
	"crypto/sha256"
)

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha256.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}
