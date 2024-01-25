package cisco_parser

import (
	"reflect"
	"net/netip"
)

// CiscoInterface represents single interface from cisco configuration data
type CiscoInterface struct {
	Name        string
	Description string
	Encapsulation string
	Ip_addr     string
	Subnet      string
	subnetRaw	netip.Prefix `csv:"skip"`
	Vrf         string
	Mtu         string
	ACLin       string
	ACLout      string
}

// newCiscoInterface is a constructor for CiscoInterface object
func newCiscoInterface(name string) *CiscoInterface {
	return &CiscoInterface{Name: name}
}

// getFields returns all "CiscoInterface" struct field names, excempt those which 
// are tagged with `csv:"skip"`
func(c *CiscoInterface) getFileds() []string {
	fields := reflect.VisibleFields(reflect.TypeOf(*c))
	result := []string{}
	for _, v := range fields {
		if v.Tag.Get("csv") == "skip" {
			continue
		}
		result = append(result, v.Name)
	}
	return result
}

// getValues returns all "CiscoInterface" struct field values, excempt those which 
// are tagged with `csv:"skip"`
func(c *CiscoInterface) getValues() []string {
	result := []string{}

	e := reflect.ValueOf(c).Elem()
	for i:= 0; i < e.NumField(); i++ {
		if e.Type().Field(i).Tag.Get("csv") == "skip" {
			continue
		}
		value := e.Field(i).Interface().(string)
		result = append(result, value)
	}
	return result
}