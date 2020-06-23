package wimatrix

import (
	"encoding/json"
	"fmt"
	"image/color"
	"time"
)

func (d *Device) publishMQ(topic string, data []byte) {
	log.Debug("Sending to %s: %s", topic, string(data))
	tkn := d.mq.Publish(topic, 0, false, data)
	if !tkn.WaitTimeout(time.Second) {
		log.Error("Error publishing message to %s: %s", topic, tkn.Error())
		return
	}
}

func (d *Device) setMode(mode Mode) {
	log.Info("Setting mode to %s", mode)
	s := fmt.Sprintf("%d", mode)
	topic := d.name + MQTTWiMatrixSetMode

	d.publishMQ(topic, []byte(s))
}

func (d *Device) setTextBrightness(brightness float32) {
	if brightness < 0 {
		brightness = 0
	}
	if brightness > 1 {
		brightness = 1
	}

	log.Info("Setting text brightness to %f", brightness)
	s := fmt.Sprintf("%f", brightness)
	topic := d.name + MQTTWiMatrixSetBrightness

	d.lastBrightness = brightness

	d.publishMQ(topic, []byte(s))
}

func (d *Device) setBGBrightness(brightness float32) {
	if brightness < 0 {
		brightness = 0
	}
	if brightness > 1 {
		brightness = 1
	}

	log.Info("Setting background brightness to %f", brightness)
	s := fmt.Sprintf("%f", brightness)
	topic := d.name + MQTTWiMatrixSetBGBrightness

	d.lastBgBrightness = brightness

	d.publishMQ(topic, []byte(s))
}

func (d *Device) setTextColor(c color.Color) {
	d.lastColor = c

	topic := d.name + MQTTWiMatrixSetTextColor

	r, g, b, _ := d.lastColor.RGBA()

	data := map[string]interface{}{
		"r": r & 0xFF,
		"g": g & 0xFF,
		"b": b & 0xFF,
	}

	dataBytes, _ := json.Marshal(data)
	d.publishMQ(topic, dataBytes)
}

func (d *Device) setBGColor(c color.Color) {
	d.lastBGColor = c

	topic := d.name + MQTTWiMatrixSetBGColor

	r, g, b, _ := d.lastBGColor.RGBA()

	data := map[string]interface{}{
		"r": r & 0xFF,
		"g": g & 0xFF,
		"b": b & 0xFF,
	}

	dataBytes, _ := json.Marshal(data)
	d.publishMQ(topic, dataBytes)
}

func (d *Device) msg(message string) {
	log.Info("Sending message: %s", message)
	topic := d.name + MQTTWimatrixMsg

	r, g, b, _ := d.lastColor.RGBA()

	data := map[string]interface{}{
		"msg": message,
		"r":   r & 0xFF,
		"g":   g & 0xFF,
		"b":   b & 0xFF,
	}

	dataBytes, _ := json.Marshal(data)

	d.publishMQ(topic, dataBytes)
}

func (d *Device) setSpeed(speed int) {
	log.Debug("Setting speed to %d", speed)

	topic := d.name + MQTTWiMatrixSetSpeed

	d.publishMQ(topic, []byte(fmt.Sprintf("%d", speed)))
}

func (d *Device) setLight() {
	log.Debug("Setting light")

	topic := MQTTSetRoomLight
	d.publishMQ(topic, []byte("1"))
	time.Sleep(time.Millisecond * 10) // Simulate button hit
	d.publishMQ(topic, []byte("0"))
}
