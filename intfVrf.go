package cisco_parser

import (
	"fmt"
	"strings"
)

// intfVrf describes interface/vrf combination, which belongs to certain subnet
type intfVrf struct {
	intf string
	vrf string
}

// newInterfaceVrf is a constructor for intfVrf.
func newInterfaceVrf(intf string) *intfVrf {
	return &intfVrf{
		intf: intf,
	}
}

// String representation of intfVrf
func(i *intfVrf) String() string {
	return fmt.Sprintf("interface: %q, vrf: %q", i.intf, i.vrf)
}

// addVrf adds VRF info to intfVrf object
func(i *intfVrf) addVrf(v string) {
	i.vrf = v
}

// intfVrfList represents all intfVrf objects, which belong to same subnet
type IntfVrfList struct {
	items []*intfVrf
}

// newInterfaceVrfList is a constructor for intfVrfList
func newInterfaceVrfList() *IntfVrfList {
	return &IntfVrfList{
		items: make([]*intfVrf, 0),
	}
}

// addItem adds intfVrf to intfVrfList
func(il *IntfVrfList) addItem(i *intfVrf) {
	il.items = append(il.items, i)
}

// String representation of intfVrfList
func(il *IntfVrfList) String() string {
	result := []string{}
	for _, v := range il.items {
		result = append(result, v.String())
	}
	return strings.Join(result, "\n")
}