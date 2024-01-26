package cisco_parser

import (
	"net/netip"
	"fmt"
)

// will be used in compare subnets feature
type subnetVrf struct {
	subnet netip.Prefix
	vrf string
}

func(s *subnetVrf) String() string {
	return fmt.Sprintf("subnet: %q, vrf: %q", s.subnet, s.vrf)
}

func newSubnetVrf(s netip.Prefix) *subnetVrf {
	return &subnetVrf{
		subnet: s,
	}
}

func(s *subnetVrf) addVrf(v string) {
	s.vrf = v
}