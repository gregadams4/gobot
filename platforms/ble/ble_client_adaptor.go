package ble

import (
	"context"
	"log"
	"strings"
	"time"

	"gobot.io/x/gobot"

	blelib "github.com/currantlabs/ble"
	"github.com/pkg/errors"
)

// newBLEDevice is constructor about blelib HCI device connection
func newBLEDevice(impl string) (d blelib.Device, err error) {
	return defaultDevice(impl)
}

// ClientAdaptor represents a Client Connection to a BLE Peripheral
type ClientAdaptor struct {
	name    string
	address string

	addr    blelib.Addr
	device  blelib.Device
	client  blelib.Client
	profile *blelib.Profile

	connected bool
	ready     chan struct{}
}

// NewClientAdaptor returns a new ClientAdaptor given an address or peripheral name
func NewClientAdaptor(address string) *ClientAdaptor {
	return &ClientAdaptor{
		name:      gobot.DefaultName("BLEClient"),
		address:   address,
		connected: false,
	}
}

// Name returns the name for the adaptor
func (b *ClientAdaptor) Name() string { return b.name }

// SetName sets the name for the adaptor
func (b *ClientAdaptor) SetName(n string) { b.name = n }

// Address returns the Bluetooth LE address for the adaptor
func (b *ClientAdaptor) Address() string { return b.address }

// Connect initiates a connection to the BLE peripheral. Returns true on successful connection.
func (b *ClientAdaptor) Connect() (err error) {
	d, err := newBLEDevice("default")
	if err != nil {
		return errors.Wrap(err, "can't new device")
	}
	blelib.SetDefaultDevice(d)
	b.device = d

	var cln blelib.Client

	ctx := blelib.WithSigHandler(context.WithTimeout(context.Background(), 3*time.Second))
	cln, err = blelib.Connect(ctx, filter(b.Address()))
	if err != nil {
		return errors.Wrap(err, "can't connect")
	}

	b.addr = cln.Address()
	b.address = cln.Address().String()
	b.SetName(cln.Name())
	b.client = cln

	p, err := b.client.DiscoverProfile(true)
	if err != nil {
		return errors.Wrap(err, "can't discover profile")
	}

	b.profile = p
	b.connected = true
	return
}

// Reconnect attempts to reconnect to the BLE peripheral. If it has an active connection
// it will first close that connection and then establish a new connection.
// Returns true on Successful reconnection
func (b *ClientAdaptor) Reconnect() (err error) {
	if b.connected {
		b.Disconnect()
	}
	return b.Connect()
}

// Disconnect terminates the connection to the BLE peripheral. Returns true on successful disconnect.
func (b *ClientAdaptor) Disconnect() (err error) {
	b.client.CancelConnection()
	return
}

// Finalize finalizes the BLEAdaptor
func (b *ClientAdaptor) Finalize() (err error) {
	return b.Disconnect()
}

// ReadCharacteristic returns bytes from the BLE device for the
// requested characteristic uuid
func (b *ClientAdaptor) ReadCharacteristic(cUUID string) (data []byte, err error) {
	if !b.connected {
		log.Fatalf("Cannot read from BLE device until connected")
		return
	}

	uuid, _ := blelib.Parse(cUUID)

	if u := b.profile.Find(blelib.NewCharacteristic(uuid)); u != nil {
		data, err = b.client.ReadCharacteristic(u.(*blelib.Characteristic))
	}

	return
}

// WriteCharacteristic writes bytes to the BLE device for the
// requested service and characteristic
func (b *ClientAdaptor) WriteCharacteristic(cUUID string, data []byte) (err error) {
	if !b.connected {
		log.Fatalf("Cannot write to BLE device until connected")
		return
	}

	uuid, _ := blelib.Parse(cUUID)

	if u := b.profile.Find(blelib.NewCharacteristic(uuid)); u != nil {
		err = b.client.WriteCharacteristic(u.(*blelib.Characteristic), data, true)
	}

	return
}

// Subscribe subscribes to notifications from the BLE device for the
// requested service and characteristic
func (b *ClientAdaptor) Subscribe(cUUID string, f func([]byte, error)) (err error) {
	if !b.connected {
		log.Fatalf("Cannot subscribe to BLE device until connected")
		return
	}

	uuid, _ := blelib.Parse(cUUID)

	if u := b.profile.Find(blelib.NewCharacteristic(uuid)); u != nil {
		h := func(req []byte) { f(req, nil) }
		err = b.client.Subscribe(u.(*blelib.Characteristic), false, h)
		if err != nil {
			return err
		}
		return nil
	}

	return
}

func filter(name string) blelib.AdvFilter {
	return func(a blelib.Advertisement) bool {
		return strings.ToLower(a.LocalName()) == strings.ToLower(name) ||
			a.Address().String() == strings.ToLower(name)
	}
}
