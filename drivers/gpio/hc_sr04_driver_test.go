package gpio

import (
	"errors"
	"strings"
	"testing"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

var _ gobot.Driver = (*HCSR04Driver)(nil)

func initTestHCSR04Driver() *HCSR04Driver {
	a := newGpioTestAdaptor()
	b := newGpioTestAdaptor()
	a.testAdaptorDigitalWrite = func() (err error) {
		return nil
	}
	b.testAdaptorDigitalRead = func() (n int, err error) {
		return 1, nil
	}
	return NewHCSR04Driver(a, b, "1", "2")
}

func TestHCSR04Driver(t *testing.T) {
	a := newGpioTestAdaptor()
	b := newGpioTestAdaptor()
	d := NewHCSR04Driver(a, b, "1", "2")

	gobottest.Assert(t, d.Pin(), "2")
	gobottest.Assert(t, d.InputPin(), "1")
	gobottest.Refute(t, d.Connection(), nil)
	gobottest.Refute(t, d.InputConnection(), nil)

	a.testAdaptorDigitalWrite = func() (err error) {
		return errors.New("write error")
	}

}

func TestHCSR04DriverStart(t *testing.T) {
	d := initTestHCSR04Driver()
	gobottest.Assert(t, d.Start(), nil)
}

func TestHCSR04DriverHalt(t *testing.T) {
	d := initTestHCSR04Driver()
	gobottest.Assert(t, d.Halt(), nil)
}

func TestHCSR04GetDistance(t *testing.T) {
	a := newGpioTestAdaptor()
	b := newGpioTestAdaptor()
	// a.testAdaptorDigitalWrite = func() (err error) {
	// 	return nil
	// }
	// b.testAdaptorDigitalRead = func() (n int, err error) {
	// 	return 1, nil
	// }
	d := NewHCSR04Driver(a, b, "1", "2")

	// go func(a *gpioTestAdaptor) {
	// 	time.Sleep(10 * time.Millisecond)
	// 	a.DigitalWrite("1", 1)
	// 	time.Sleep(10 * time.Millisecond)
	// 	a.DigitalWrite("1", 0)
	// }(a)

	// err := d.GetDistance()
	// gobottest.Assert(t, err, nil)
	// log.Println(dist.CM())

	err := d.GetDistance()
	gobottest.Assert(t, err, ErrHCSR04ReadTimeout)

	err = d.GetDistanceSample(10, 60*time.Millisecond)
	gobottest.Assert(t, err, ErrHCSR04ReadTimeout)
}

// func TestLedDriverToggle(t *testing.T) {
// 	d := initTestLedDriver()
// 	d.Off()
// 	d.Toggle()
// 	gobottest.Assert(t, d.State(), true)
// 	d.Toggle()
// 	gobottest.Assert(t, d.State(), false)
// }
//
// func TestLedDriverBrightness(t *testing.T) {
// 	a := newGpioTestAdaptor()
// 	d := NewLedDriver(a, "1")
// 	a.testAdaptorPwmWrite = func() (err error) {
// 		err = errors.New("pwm error")
// 		return
// 	}
// 	gobottest.Assert(t, d.Brightness(150), errors.New("pwm error"))
// }
//
func TestHCSR04DriverDefaultName(t *testing.T) {
	d := initTestHCSR04Driver()
	gobottest.Assert(t, strings.HasPrefix(d.Name(), "HC-SR04"), true)
}

func TestHCSR04DriverSetName(t *testing.T) {
	d := initTestHCSR04Driver()
	d.SetName("distance")
	gobottest.Assert(t, strings.HasPrefix(d.Name(), "distance"), true)
}
