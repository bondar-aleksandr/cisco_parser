package cisco_parser

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"errors"
)

// slice of supported output formats
var formatSupported = []string{"csv", "json"}

var ErrUnsupportedOutputFormat = errors.New("output format not supported")

// Serializer represents object for *Device serialization.
type Serializer struct {
	destination io.Writer
	Device *Device
	format string
}

// NewSerializer is constructor, returns instance of *Serializer
// with fields specified.
func NewSerializer(w io.Writer, d *Device, f string) (*Serializer, error) {
	if !slices.Contains(formatSupported, f) {
		return nil, ErrUnsupportedOutputFormat
	}
	return &Serializer{
		destination: w,
		Device: d,
		format: f,
	}, nil
}

// Serialize serializes data to s.destination based on output format.
func(s *Serializer) Serialize() error {
	switch s.format {
	case "csv":
		return s.toCSV()
	case "json":
		return s.toJSON()
	default:
		return nil
	}
}

// ToJSON writes device json-formatted data to serializer destination.
func(s *Serializer) toJSON() error { 
	if !s.Device.parsed {
		if err := s.Device.parse(); err != nil {
			return fmt.Errorf("can't serialize: %w", err)
		}
	}
	json_data, _ := json.MarshalIndent(s.Device, "", "  ")
	_, err := s.destination.Write(json_data)
	if err != nil {
		errorLogger.Println("Unable to write json data because of:", err.Error())
		return fmt.Errorf("can't serialize: %w", err)
	}
	infoLogger.Println("Writing JSON data done")
	return nil
}


// ToCSV writes device CSV-formatted data to serializer destination.
func (s *Serializer) toCSV() error {
	cw := csv.NewWriter(s.destination)
	headers, err := s.Device.getIntfFields()
	if err != nil {
		return fmt.Errorf("can't serialize: %w", err)
	}
	cw.Write(headers)

	interfaces, err := s.Device.getIntfNames()
	if err != nil {
		return fmt.Errorf("can't serialize: %w", err)
	}

	for _, intf := range interfaces {
		line := s.Device.Interfaces[intf].getValues()
		cw.Write(line)
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		errorLogger.Println("Unable to write csv data because of:", err.Error())
		return fmt.Errorf("can't serialize: %w", err)
	}
	infoLogger.Println("Writing CSV data done")
	return nil
}