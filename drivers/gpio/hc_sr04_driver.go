package gpio

import (
	"errors"
	"math"
	"sync"
	"time"

	"gobot.io/x/gobot"
)

var (
	//ErrHCSR04ReadTimeout should be returned if a read takes too long to respond (currently 1 second)
	ErrHCSR04ReadTimeout = errors.New("read request timed out")
)

//HCSR04Driver contains commands for an HC-SR04 ultrasonic range sensor
type HCSR04Driver struct {
	inputPin   string
	outputPin  string
	name       string
	inputConn  DigitalReader
	outputConn DigitalWriter
	timeout    time.Duration
	gobot.Eventer
	gobot.Commander
	lock     *sync.Mutex
	Distance hcsr04Distance
}

//NewHCSR04Driver will return a driver for an HC-SR04 ultrasonic range sensor
func NewHCSR04Driver(outputConn DigitalWriter, inputConn DigitalReader, inputPin, outputPin string, v ...time.Duration) *HCSR04Driver {
	// func NewHCSR04Driver(input DigitalReader, output DigitalWriter, inputPin, outputPin string, v ...time.Duration) *HCSR04Driver {
	h := &HCSR04Driver{
		inputPin:   inputPin,
		outputPin:  outputPin,
		name:       gobot.DefaultName("HC-SR04"),
		inputConn:  inputConn,
		outputConn: outputConn,
		timeout:    40 * time.Millisecond,
		Eventer:    gobot.NewEventer(),
		Commander:  gobot.NewCommander(),
		lock:       &sync.Mutex{},
	}

	return h
}

// Start implements the Driver interface
func (h *HCSR04Driver) Start() (err error) { return }

// Halt implements the Driver interface
func (h *HCSR04Driver) Halt() (err error) { return }

func (h *HCSR04Driver) Name() string {
	return h.name
}

func (h *HCSR04Driver) SetName(name string) {
	h.name = name
}

func (h *HCSR04Driver) InputPin() string {
	return h.inputPin
}

func (h *HCSR04Driver) Pin() string {
	return h.outputPin
}

func (h *HCSR04Driver) Connection() gobot.Connection {
	return h.outputConn.(gobot.Connection)
}

func (h *HCSR04Driver) InputConnection() gobot.Connection {
	return h.inputConn.(gobot.Connection)
}

func (h *HCSR04Driver) GetDistanceSample(samples int, intervals ...time.Duration) error {
	var distances []hcsr04Distance
	var err error
	var errCount int

	interval := 60 * time.Millisecond
	if len(intervals) > 0 {
		interval = intervals[0]
	}

	for i := 0; i < samples; i++ {
		err = h.GetDistance()
		if err != nil {
			if errCount > samples/2 {
				return ErrHCSR04ReadTimeout
			}
			errCount++
			continue
		}

		distances = append(distances, h.Distance)
		time.Sleep(interval)
	}

	// for _, dist := range distances {
	// 	log.Println(dist.CM())
	// }

	h.Distance = distances[int(math.Floor(float64(len(distances)/2)))]
	return nil
}

func (h *HCSR04Driver) GetDistance() error {
	h.lock.Lock()
	var start time.Time
	var value int
	var err error
	var set bool

	if err = h.outputConn.DigitalWrite(h.Pin(), 0); err != nil {
		h.lock.Unlock()
		return err
	}

	time.Sleep(5 * time.Nanosecond)

	if err = h.outputConn.DigitalWrite(h.Pin(), 1); err != nil {
		h.lock.Unlock()
		return err
	}

	absStart := time.Now()

	time.Sleep(10 * time.Nanosecond)

	if err = h.outputConn.DigitalWrite(h.Pin(), 0); err != nil {
		h.lock.Unlock()
		return err
	}

	for {
		value, err = h.inputConn.DigitalRead(h.InputPin())
		if err != nil {
			break
		}

		if !set {
			if value == 1 {
				start = time.Now()
				set = true
			}
		} else {
			if value == 0 {
				h.Distance.duration = time.Since(start)
				break
			}
		}

		if time.Since(absStart) >= h.timeout {
			h.lock.Unlock()
			return ErrHCSR04ReadTimeout
		}

	}

	h.lock.Unlock()
	return nil
}

type hcsr04Distance struct {
	duration time.Duration
}

func (h *hcsr04Distance) Duration() time.Duration {
	return h.duration
}

func (h *hcsr04Distance) CM() float64 {
	return (h.duration.Seconds() * 34300) / 2
}
