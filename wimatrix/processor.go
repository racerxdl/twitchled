package wimatrix

import (
	"fmt"
	"golang.org/x/image/colornames"
	"time"
)

func (d *Device) processEvent(e event) {
	switch e.GetType() {
	case eventNewSub:
		d.processNewSub(e.(*newSubEvent))
	case eventNewFollower:
		d.processNewFollower(e.(*newFollowerEvent))
	case eventBGColor:
		d.processBGColor(e.(*bgColorEvent))
	case eventTextColor:
		d.processTextColor(e.(*textColorEvent))
	case eventMessage:
		d.processMessage(e.(*messageEvent))
	case eventBGBrightness:
		d.processBGBrightness(e.(*bgBrightnessEvent))
	case eventTextBrightness:
		d.processTextBrightness(e.(*textBrightnessEvent))
	case eventNewMode:
		d.processNewMode(e.(*newModeEvent))
	case eventSetSpeed:
		d.processSetSpeed(e.(*newSetSpeedEvent))
	case eventSetLight:
		d.processSetLight(e.(*newSetLightEvent))
	default:
		log.Error("Unknown event type: (%s) %d", e.GetType(), e.GetType())
	}
}

func (d *Device) processNewSub(e *newSubEvent) {
	m := d.currentMode
	bgc := d.lastBGColor
	txc := d.lastColor

	d.setMode(ModeBackgroundStringDisplay)
	d.setBGColor(colornames.Teal)
	d.setTextColor(colornames.Green)

	d.msg(fmt.Sprintf("%s TKS SUB %d MESES!", e.username, e.months))
	// TODO: Effects
	time.Sleep(time.Second * 10)

	d.setMode(m)
	d.setBGColor(bgc)
	d.setTextColor(txc)
}

func (d *Device) processNewFollower(e *newFollowerEvent) {
	m := d.currentMode
	bgc := d.lastBGColor
	txc := d.lastColor

	d.setMode(ModeBackgroundStringDisplay)
	d.setBGColor(colornames.Teal)
	d.setTextColor(colornames.Green)

	d.msg(fmt.Sprintf("%s TKS FOLLOW!", e.username))
	// TODO: Effects
	time.Sleep(time.Second * 10)

	d.setMode(m)
	d.setBGColor(bgc)
	d.setTextColor(txc)
}

func (d *Device) processBGColor(e *bgColorEvent) {
	d.setBGColor(e.color)
}

func (d *Device) processTextColor(e *textColorEvent) {
	d.setTextColor(e.color)
}

func (d *Device) processMessage(e *messageEvent) {
	d.msg(e.text)
	time.Sleep(time.Second * 5) // Wait before returning to event loop
}

func (d *Device) processBGBrightness(e *bgBrightnessEvent) {
	if e.brightness > 0.2 {
		e.brightness = 0.2
	}
	d.setBGBrightness(e.brightness)
}

func (d *Device) processTextBrightness(e *textBrightnessEvent) {
	d.setTextBrightness(e.brightness)
}

func (d *Device) processNewMode(e *newModeEvent) {
	d.setMode(e.mode)
}

func (d *Device) processSetSpeed(e *newSetSpeedEvent) {
	d.setSpeed(e.speed)
}

func (d *Device) processSetLight(e *newSetLightEvent) {
	d.setLight()
}
