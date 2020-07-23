package wimatrix

import "fmt"

type Mode uint8

const (
	ModeStringDisplay           Mode = 0
	ModeBackgroundOnly          Mode = 1
	ModeBackgroundStringDisplay Mode = 2
	ModeClock                   Mode = 3
	ModeBackgroundClock         Mode = 4
)

var modeNames = map[Mode]string{
	ModeStringDisplay:           "String Display",
	ModeBackgroundOnly:          "Background Only",
	ModeBackgroundStringDisplay: "String Display with Background",
	ModeClock:                   "Clock",
	ModeBackgroundClock:         "Clock with Background",
}

var Modes = []Mode{
	ModeStringDisplay          ,
	ModeBackgroundOnly,
	ModeBackgroundStringDisplay,
	ModeClock,
	ModeBackgroundClock,
}

func (m Mode) String() string {
	v, ok := modeNames[m]
	if !ok {
		return fmt.Sprintf("Unknown Mode (%d)", m)
	}

	return v
}
