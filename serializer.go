package cisco_parser

import (
	"encoding/json"
	"encoding/csv"
	"io"
	"fmt"
)

// Serializer represents object for *Device serialization.
type Serializer struct {
	destination io.Writer
	Device *Device
}

// NewSerializer is constructor, returns instance of *Serializer
// with fields specified
func NewSerializer(w io.Writer, d *Device) *Serializer {
	return &Serializer{
		destination: w,
		Device: d,
	}
}

// ToJSON writes device json-formatted data to serializer destination.
func(s *Serializer) ToJSON() error { 
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
func (s *Serializer) ToCSV() error {
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