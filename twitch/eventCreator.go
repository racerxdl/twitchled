package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/racerxdl/twitchled/twitch/twitchdata"
)

func makeSubscribeEvent(channelId, data string) (ChatEvent, error) {
	bitsData := twitchdata.ChannelSubscribeMessageData{}

	err := json.Unmarshal([]byte(data), &bitsData)
	if err != nil {
		return nil, err
	}

	return MakeSubscribeEventData(channelId, bitsData), nil
}
func makeRewardRedeemed(channelId, data string) (ChatEvent, error) {
	redemptionData := twitchdata.RedemptionData{}
	err := json.Unmarshal([]byte(data), &redemptionData)
	if err != nil {
		return nil, err
	}

	return MakeRewardRedemptionEventData(channelId, redemptionData), nil
}

func makePointsEvent(channelId, data string) (ChatEvent, error) {
	tmp := make(map[string]interface{})

	err := json.Unmarshal([]byte(data), &tmp)

	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(tmp["data"].(map[string]interface{})["redemption"])

	if err != nil {
		return nil, err
	}

	pointType := tmp["type"].(string)

	switch pointType {
	case twitchdata.ChannelPointsRewardRedeemed:
		return makeRewardRedeemed(channelId, string(content))
	}

	// if not know, return error
	return nil, fmt.Errorf("unknown point reward type: %s", pointType)
}

func makeBitsEvent(channelId, data string) (ChatEvent, error) {
	bitsData := twitchdata.BitEventsV2{}

	err := json.Unmarshal([]byte(data), &bitsData)
	if err != nil {
		return nil, err
	}

	return MakeBitsV2EventData(channelId, bitsData), nil
}
