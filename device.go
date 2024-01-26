package cisco_parser

import (
	"errors"
	"fmt"
	"io"
	"log"
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
	Interfaces map[string]*CiscoInterface
	subnets    map[*subnetVrf]string
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
		subnets:    make(map[*subnetVrf]string),
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

func (d *Device) addSubnetVrf(s *subnetVrf, intf string) error {
	_, exists := d.subnets[s]
	if !exists {
		d.subnets[s] = intf
		return nil
	} else {
		return ErrDublicateInterface
	}
}

func (d *Device) GetSubnets() (string, error) {
	if !d.parsed {
		if err := d.parse(); err != nil {
			return "", fmt.Errorf("can't get interfaces: %w", err)
		}
	}
	result := strings.Builder{}
	for k, v := range d.subnets {
		line := fmt.Sprintf("%s interface: %s\n", k.String(), v)
		result.WriteString(line)
	}
	return result.String(), nil
}
