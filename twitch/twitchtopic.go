package twitch

import "strings"

const (
	// Bits Event v2 Message
	topicChannelBitsEventv2 = "channel-bits-events-v2"
	// Channel Points Event Message
	topicChannelPoints = "channel-points-channel-v1"
	// Channel Subscribe Event Message
	topicChannelSubscribe = "channel-subscribe-events-v1"
)

// channelIdAndEventFromTopic returns the event and channelId on the specified topic
// returns empty string if unknown event
func channelIdAndEventFromTopic(topic string) (channelId, event string) {
	v := strings.Split(topic, ".")

	if len(v) != 2 { // Should have exactly two items
		return "", ""
	}

	return v[1], v[0]
}
