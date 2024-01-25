package cisco_parser

import (
	// "encoding/json"
	"encoding/csv"
	"io"
	"strings"
	"path/filepath"
	"fmt"
)

type Serializer struct {
	destination io.Writer
	device *Device
}

func NewSerializer(w io.Writer, d *Device) *Serializer {
	return &Serializer{
		destination: w,
		device: d,
	}
}

// func (c CiscoInterfaceMap) ToJSON(w io.Writer) { // For testing purpose, to get structured data to deserialize from
// 	json_data, _ := json.MarshalIndent(c, "", "  ")
// 	_, err := w.Write(json_data)
// 	if err != nil {
// 		errorLogger.Println("Unable to write json data because of:", err.Error())
// 	}
// 	infoLogger.Println("Writing JSON data done")
// }


// ToCSV writes device CSV-formatted data to serializer destination
func (s *Serializer) ToCSV() error {
	cw := csv.NewWriter(s.destination)
	headers, err := s.device.getIntfFields()
	if err != nil {
		return fmt.Errorf("can't get interface fields: %w", err)
	}
	cw.Write(headers)

	interfaces, err := s.device.getIntfNames()
	if err != nil {
		return fmt.Errorf("can't get interfaces names: %w", err)
	}

	for _, intf := range interfaces {
		line := s.device.interfaces[intf].getValues()
		cw.Write(line)
	}
	cw.Flush()
	infoLogger.Println("Writing CSV data done")
	return nil
}

func FileExtReplace(f string, ex string) string {
	bareName := strings.TrimSuffix(f, filepath.Ext(f))
	return fmt.Sprintf("%s.%s", bareName, ex)
}