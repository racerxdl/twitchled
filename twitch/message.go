package twitch

import (
	"github.com/google/uuid"
	"time"
)

const maxUUIDRetryCount = 1024 // This should never happen

// makeUniqueId generates a random UUID for message nonce
func makeUniqueId() string {
	// uuid might fail sometimes so we try until we get something
	// this can only happen if /dev/urandom returns a Try again
	// that's why we wait one millis
	msgId, err := uuid.NewUUID()

	retries := 0

	for err != nil && retries > maxUUIDRetryCount {
		time.Sleep(time.Millisecond)
		msgId, err = uuid.NewUUID()
		retries++
	}

	if retries == maxUUIDRetryCount {
		log.Error("ERROR: MAX UUID GENERATOR RETRY COUNT REACHED! SERVER IS BUGGY")
	}

	return msgId.String()
}

// makeMessage creates a twitch websocket message with the specified type and data
func makeMessage(msgType string, data map[string]interface{}) (msgId string, message map[string]interface{}) {
	message = map[string]interface{}{
		"type":  msgType,
		"nonce": makeUniqueId(),
		"data":  data,
	}

	return message["nonce"].(string), message
}
