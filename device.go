package cisco_parser

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"slices"
	"strings"
)

const (
	ios  = "ios"
	nxos = "nxos"
)

var (
	infoLogger  *log.Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger  *log.Logger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger *log.Logger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

var (
	ErrDublicateInterface = errors.New("dublicate interface in device")
	ErrPlatformUnknown    = errors.New("platform unknown")
)

// Device is aggregate type, includes interfaces and subnets(TBD)
type Device struct {
	source     io.Reader
	platform   string
	parsed     bool
	Hostname   string
	Interfaces map[string]*CiscoInterface
	subnets    map[netip.Prefix]*intfVrfList
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
		source:     s,
		platform:   platform,
		parsed:     false,
		Interfaces: make(map[string]*CiscoInterface),
		subnets:    make(map[netip.Prefix]*intfVrfList),
	}, nil
}

// addInterface adds CiscoInterface object to device.interfaces structure.
// Returns error if interface with the same name already there
func (d *Device) addInterface(intf *CiscoInterface) error {
	_, exists := d.Interfaces[intf.Name]
	if !exists {
		d.Interfaces[intf.Name] = intf
		return nil
	} else {
		return ErrDublicateInterface
	}
}

// getIntfFields returns slice of Ciscointerface struct's field names and error if any.
func (d *Device) getIntfFields() ([]string, error) {
	result := []string{}
	if !d.parsed {
		if err := d.parse(); err != nil {
			return result, fmt.Errorf("can't get interface fields: %w", err)
		}
	}
	for _, v := range d.Interfaces {
		result = v.getFileds()
		break
	}
	return result, nil
}

// getIntfNames returns ascending ordered slice of device's interface names and error if any
func (d *Device) getIntfNames() ([]string, error) {
	result := []string{}
	if !d.parsed {
		if err := d.parse(); err != nil {
			return result, fmt.Errorf("can't get interfaces names: %w", err)
		}
	}
	for k := range d.Interfaces {
		result = append(result, k)
	}
	slices.Sort(result)
	return result, nil
}

// intfAmount returns amount of parsed interfaces
func (d *Device) intfAmount() int {
	return len(d.Interfaces)
}

// addSubnet adds subnet to device. The function checks whether subnet
// already exists, and if so, adds *intfVrf object to the subnet corresponding intfVrfList.
// Otherwise, if subnet not listed in device.subnets, the function creates intfVrfList
// and adds subnet with newly created intfVrfList to device.subnets
func (d *Device) addSubnet(p netip.Prefix, i *intfVrf) {
	_, exists := d.subnets[p]
	if !exists {
		il := newInterfaceVrfList()
		il.addItem(i)
		d.subnets[p] = il
	} else {
		il := d.subnets[p]
		il.addItem(i)
	}
}

// For testing purposes
func (d *Device) GetSubnets() (string, error) {
	if !d.parsed {
		if err := d.parse(); err != nil {
			return "", fmt.Errorf("can't get interfaces: %w", err)
		}
	}
	result := strings.Builder{}
	for k, v := range d.subnets {
		line := fmt.Sprintf("\tsubnet: %q\n%s", k, v.String())
		result.WriteString(line)
	}
	return result.String(), nil
}

func (d *Device) GetSubnet(p netip.Prefix) *intfVrfList {
	_, exists := d.subnets[p]
	if !exists {
		return nil
	} else {
		return d.subnets[p]
	}
}
