package wimatrix

import (
	"github.com/asaskevich/EventBus"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/foize/go.fifo"
	"github.com/quan-to/slog"
	"golang.org/x/image/colornames"
	"image/color"
	"time"
)

var log = slog.Scope("WiMatrix")

type Device struct {
	name        string
	mq          mqtt.Client
	ev          EventBus.Bus
	lastColor   color.Color
	lastBGColor color.Color
	eventQueue  *fifo.Queue
	running     bool
	currentMode Mode
}

func MakeWiiMatrix(name string, mq mqtt.Client, ev EventBus.Bus) *Device {
	return &Device{
		name:        name,
		mq:          mq,
		ev:          ev,
		lastColor:   colornames.White,
		lastBGColor: colornames.Black,
		eventQueue:  fifo.NewQueue(),
		running:     false,
		currentMode: ModeClock,
	}
}

func (d *Device) Start() {
	d.subEventBus()
	d.running = true
	go d.eventLoop()
}

func (d *Device) Stop() {
	d.running = false
	d.unSubEventBus()
}

func (d *Device) putEvent(event event) {
	d.eventQueue.Add(event)
}

func (d *Device) eventLoop() {
	for d.running {
		rawe := d.eventQueue.Next()
		if rawe != nil {
			e, ok := rawe.(event)
			if ok && !e.Expired() {
				d.processEvent(e)
			}
		}

		time.Sleep(time.Millisecond)
	}
}
