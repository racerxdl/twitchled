package twitch

type User struct {
	Id          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

type RewardData struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type RedemptionData struct {
	Id        string     `json:"id"`
	User      User       `json:"user"`
	Reward    RewardData `json:"reward"`
	UserInput string     `json:"user_input"`
	Status    string     `json:"status"`
}

/*
    {
    "type": "reward-redeemed",
    "data": {
        "timestamp": "2020-05-29T02:06:43.063269581Z",
        "redemption": {
            "id": "d3444e0d-49d8-4f98-8e34-932344204136",
            "user": {
                "id": "44043625",
                "login": "racerxdl",
                "display_name": "RacerXDL"
            },
            "channel_id": "44043625",
            "redeemed_at": "2020-05-29T02:06:42.985912724Z",
            "reward": {
                "id": "b1834989-c3ac-4d78-9888-180c5768f0e4",
                "channel_id": "44043625",
                "title": "HUEPainel",
                "prompt": "Mostre uma mensagem no painel de LED",
                "cost": 1,
                "is_user_input_required": true,
                "is_sub_only": false,
                "image": null,
                "default_image": {
                    "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/default-1.png",
                    "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/default-2.png",
                    "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/default-4.png"
                },
                "background_color": "#E3A300",
                "is_enabled": true,
                "is_paused": false,
                "is_in_stock": true,
                "max_per_stream": {
                    "is_enabled": false,
                    "max_per_stream": 0
                },
                "should_redemptions_skip_request_queue": false,
                "template_id": null,
                "updated_for_indicator_at": "2020-05-29T00:21:39.858284128Z"
            },
            "user_input": "iguliu",
            "status": "UNFULFILLED"
        }
    }
}
*/
