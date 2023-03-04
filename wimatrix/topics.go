package wimatrix

// MQTT Topics
const (
	MQTTWimatrixMsg             = "_msg"
	MQTTWiMatrixSetBrightness   = "_brightness"
	MQTTWiMatrixSetBGBrightness = "_bgbrightness"
	MQTTWiMatrixSetBGColor      = "_bgcolor"
	MQTTWiMatrixSetTextColor    = "_textcolor"
	MQTTWiMatrixSetMode         = "_mode"
	MQTTWiMatrixSetSpeed        = "_scrollspeed"
	MQTTSetRoomLight            = "ENTRADA/036"
)

// EventBus Topics
const (
	EvNewSub            = "WiMatrix:NewSub"
	EvNewBits           = "WiMatrix:NewBits"
	EvNewFollower       = "WiMatrix:NewFollower"
	EvNewMsg            = "WiMatrix:NewMsg"
	EvSetTextColor      = "WiMatrix:SetTextColor"
	EvSetBgColor        = "WiMatrix:SetBackgroundColor"
	EvSetTextBrightness = "WiMatrix:SetTextBrightness"
	EvSetBgBrightness   = "WiMatrix:SetBackgroundBrightness"
	EvNewMode           = "WiMatrix:SetMode"
	EvSetSpeed          = "WiMatrix:SetSpeed"
	EvSetLight          = "Room:SetLight"
)
