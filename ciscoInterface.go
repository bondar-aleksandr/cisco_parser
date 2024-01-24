package cisco_parser

import (
	"reflect"
	"net/netip"
)

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

func (c CiscoInterface) getFileds() []string {
	fields := reflect.VisibleFields(reflect.TypeOf(c))
	result := []string{}
	for _, v := range fields {
		if v.Tag.Get("csv") == "skip" {
			continue
		}
		result = append(result, v.Name)
	}
	return result
}

// TODO: add getter for all CiscoInterface values