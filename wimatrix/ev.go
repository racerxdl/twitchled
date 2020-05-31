package wimatrix

import (
	"image/color"
	"time"
)

func (d *Device) subEventBus() {
	d.ev.Subscribe(EvNewSub, d.evNewSub)
	d.ev.Subscribe(EvNewFollower, d.evNewFollower)
	d.ev.Subscribe(EvSetTextColor, d.evSetTextColor)
	d.ev.Subscribe(EvSetBgColor, d.evSetBackgroundColor)
	d.ev.Subscribe(EvNewMsg, d.evNewMessage)
	d.ev.Subscribe(EvSetTextBrightness, d.evSetTextBrightness)
	d.ev.Subscribe(EvSetBgBrightness, d.evSetBGBrightness)
	d.ev.Subscribe(EvNewMode, d.evNewMode)
}

func (d *Device) unSubEventBus() {
	d.ev.Unsubscribe(EvNewSub, d.evNewSub)
	d.ev.Unsubscribe(EvNewFollower, d.evNewFollower)
	d.ev.Unsubscribe(EvSetTextColor, d.evSetTextColor)
	d.ev.Unsubscribe(EvSetBgColor, d.evSetBackgroundColor)
	d.ev.Unsubscribe(EvNewMsg, d.evNewMessage)
	d.ev.Unsubscribe(EvSetTextBrightness, d.evSetTextBrightness)
	d.ev.Unsubscribe(EvSetBgBrightness, d.evSetBGBrightness)
	d.ev.Unsubscribe(EvNewMode, d.evNewMode)
}

func (d *Device) evNewSub(username string, months int) {
	d.eventQueue.Add(&newSubEvent{
		username: username,
		months:   months,
		when:     time.Now(),
	})
}

func (d *Device) evNewFollower(username string) {
	d.eventQueue.Add(&newFollowerEvent{
		username: username,
		when:     time.Now(),
	})
}

func (d *Device) evSetTextColor(color color.Color) {
	d.eventQueue.Add(&textColorEvent{
		color: color,
		when:  time.Now(),
	})
}

func (d *Device) evSetBackgroundColor(color color.Color) {
	d.eventQueue.Add(&textColorEvent{
		color: color,
		when:  time.Now(),
	})
}

func (d *Device) evSetTextBrightness(brightness float32) {
	d.eventQueue.Add(&textBrightnessEvent{
		brightness: brightness,
		when:       time.Now(),
	})
}

func (d *Device) evSetBGBrightness(brightness float32) {
	d.eventQueue.Add(&bgBrightnessEvent{
		brightness: brightness,
		when:       time.Now(),
	})
}

func (d *Device) evNewMessage(message string) {
	d.eventQueue.Add(&messageEvent{
		text: message,
		when: time.Now(),
	})
}

func (d *Device) evNewMode(mode Mode) {
	d.eventQueue.Add(&newModeEvent{
		mode: mode,
		when: time.Now(),
	})
}
