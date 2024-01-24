package cisco_parser

import (
	// "errors"
	"fmt"
	"io"
	"net/netip"
	"slices"
)

const (
	IOS = "ios"
	NXOS = "nxos"
)

// var ErrNoInterfaces = errors.New("no interfaces parsed")

type subnetVrf struct {
	subnet netip.Prefix
	vrf string
}

type Device struct {
	source io.Reader
	platform string
	interfaces map[string]*CiscoInterface
	subnets map[subnetVrf]*CiscoInterface
}

func NewDevice(s io.Reader, p string) (*Device, error) {
	var platform string
	switch p {
	case IOS:
		platform = IOS
	case NXOS:
		platform = NXOS
	default:
		return nil, fmt.Errorf("platform unknown")
	}
	return &Device{
		source: s,
		platform: platform,
		interfaces: make(map[string]*CiscoInterface),
		subnets: make(map[subnetVrf]*CiscoInterface),
	}, nil
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