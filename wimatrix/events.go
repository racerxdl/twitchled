package wimatrix

import (
	"image/color"
	"time"
)

type eventType int

const (
	eventMessage        eventType = iota
	eventTextColor      eventType = iota
	eventBGColor        eventType = iota
	eventTextBrightness eventType = iota
	eventBGBrightness   eventType = iota
	eventNewSub         eventType = iota
	eventNewFollower    eventType = iota
	eventNewMode        eventType = iota
	eventSetSpeed       eventType = iota
)

const expirationDuration = time.Minute * 5

type event interface {
	GetType() eventType
	Expired() bool
}

// region
type messageEvent struct {
	text string
	when time.Time
}

func (e messageEvent) GetType() eventType {
	return eventMessage
}

func (e messageEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type textColorEvent struct {
	color color.Color
	when  time.Time
}

func (e textColorEvent) GetType() eventType {
	return eventTextColor
}

func (e textColorEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type bgColorEvent struct {
	color color.Color
	when  time.Time
}

func (e bgColorEvent) GetType() eventType {
	return eventBGColor
}

func (e bgColorEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type textBrightnessEvent struct {
	brightness float32
	when       time.Time
}

func (e textBrightnessEvent) GetType() eventType {
	return eventTextBrightness
}

func (e textBrightnessEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type bgBrightnessEvent struct {
	brightness float32
	when       time.Time
}

func (e bgBrightnessEvent) GetType() eventType {
	return eventBGBrightness
}

func (e bgBrightnessEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type newSubEvent struct {
	username string
	months   int
	when     time.Time
}

func (e newSubEvent) GetType() eventType {
	return eventNewSub
}

func (e newSubEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type newFollowerEvent struct {
	username string
	when     time.Time
}

func (e newFollowerEvent) GetType() eventType {
	return eventNewFollower
}

func (e newFollowerEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type newModeEvent struct {
	mode Mode
	when time.Time
}

func (e newModeEvent) GetType() eventType {
	return eventNewMode
}

func (e newModeEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion

// region
type newSetSpeedEvent struct {
	speed int
	when  time.Time
}

func (e newSetSpeedEvent) GetType() eventType {
	return eventSetSpeed
}

func (e newSetSpeedEvent) Expired() bool {
	return e.when.Add(expirationDuration).Before(time.Now())
}

// endregion
