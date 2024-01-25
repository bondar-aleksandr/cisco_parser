package cisco_parser

import (
	"errors"
	"fmt"
	"io"
	"net/netip"
	"slices"
)

const (
	ios = "ios"
	nxos = "nxos"
)

var (
	ErrDublicateInterface = errors.New("dublicate interface in device")
	ErrPlatformUnknown = errors.New("platform unknown")
)

type subnetVrf struct {
	subnet netip.Prefix
	vrf string
}

// Device is aggregate type, includes interfaces and subnets(TBD)
type Device struct {
	source io.Reader
	platform string
	interfaces map[string]*CiscoInterface
	subnets map[subnetVrf]string
}

// NewDevice is constructor for Device object. Returns instance of Device with
// specified fields set. Returns error if platform string is unknown
func NewDevice(s io.Reader, p string) (*Device, error) {
	var platform string
	switch p {
	case ios:
		platform = ios
	case nxos:
		platform = nxos
	default:
		return nil, ErrPlatformUnknown
	}
	return &Device{
		source: s,
		platform: platform,
		interfaces: make(map[string]*CiscoInterface),
		subnets: make(map[subnetVrf]string),
	}, nil
}


// addInterface adds CiscoInterface object to device.interfaces structure.
// Returns error if interface with the same name already there
func(d *Device) addInterface(intf *CiscoInterface) error {
	_, exists := d.interfaces[intf.Name]
	if !exists {
		d.interfaces[intf.Name] = intf
		return nil
	} else {
		return ErrDublicateInterface
	}
}

// getIntfFields returns slice of Ciscointerface struct's field names and error if any.
func(d *Device) getIntfFields() ([]string, error) {
	result := []string{}
	if d.intfAmount() == 0 {
		if err := d.parse(); err != nil {
			return result, fmt.Errorf("can't parse config: %w", err)
		}
	}
	for _, v := range d.interfaces {
		result = v.getFileds()
		break
	}
	return result, nil
}

// getIntfNames returns ascending ordered slice of device's interface names and error if any
func(d *Device) getIntfNames() ([]string, error) {
	result := []string{}
	if d.intfAmount() == 0 {
		if err := d.parse(); err != nil {
			return result, fmt.Errorf("can't parse config: %w", err)
		}
	}
	for k, _ := range d.interfaces {
		result = append(result, k)
	}
	slices.Sort(result)
	return result, nil
}

// intfAmount returns amount of parsed interfaces
func(d *Device) intfAmount() int {
	return len(d.interfaces)
}