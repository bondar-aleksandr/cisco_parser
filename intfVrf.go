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
	return fmt.Sprintf("interface: %q, vrf: %q\n", i.intf, i.vrf)
}

// addVrf adds VRF info to intfVrf object
func(i *intfVrf) addVrf(v string) {
	i.vrf = v
}

// intfVrfList represents all intfVrf objects, which belong to same subnet
type intfVrfList struct {
	items []*intfVrf
}

// newInterfaceVrfList is a constructor for intfVrfList
func newInterfaceVrfList() *intfVrfList {
	return &intfVrfList{
		items: make([]*intfVrf, 0),
	}
}

// addItem adds intfVrf to intfVrfList
func(il *intfVrfList) addItem(i *intfVrf) {
	il.items = append(il.items, i)
}

// String representation of intfVrfList
func(il *intfVrfList) String() string {
	result := strings.Builder{}
	line := "--------\n"
	result.WriteString(line)
	for _, v := range il.items {
		line = v.String()
		result.WriteString(line)
	}
	line = "--------\n"
	result.WriteString(line)
	return result.String()
}