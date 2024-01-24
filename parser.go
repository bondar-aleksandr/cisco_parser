package cisco_parser

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/netip"
	"os"
	// "reflect"
	"regexp"
	// "sort"
	"strings"
)

var (
	infoLogger  *log.Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// warnLogger *log.Logger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger *log.Logger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

var ErrParsingFailed = errors.New("parsing failed")


// type CiscoInterfaceMap map[string]*CiscoInterface

// func (c CiscoInterfaceMap) GetSortedKeys() []string {
// 	keys := make([]string, 0)
// 	for k := range c {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)
// 	return keys
// }

// func (c CiscoInterfaceMap) getFields() []string {
// 	fields := reflect.VisibleFields(reflect.TypeOf(CiscoInterface{}))
// 	result := []string{}
// 	for _, v := range fields {
// 		result = append(result, v.Name)
// 	}
// 	return result
// }



const (
	INTF_REGEXP   = `^interface (\S+)`
	DESC_REGEXP   = ` {1,2}description (.*)$`
	ENCAP_REGEXP  = ` {1,2}encapsulation (.+)`
	IP_REGEXP     = ` {1,2}ip(?:v4)? address (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(?: secondary)?`
	VRF_REGEXP    = ` {1,2}vrf(?: forwarding| member)? (\S+)`
	MTU_REGEXP    = ` {1,2}(?:ip )?mtu (\S+)`
	ACLIN_REGEXP  = ` {1,2}access-group (\S+) in`
	ACLOUT_REGEXP = ` {1,2}access-group (\S+) out`
)

var (
	intf_compiled   = regexp.MustCompile(INTF_REGEXP)
	desc_compiled   = regexp.MustCompile(DESC_REGEXP)
	encap_compiled  = regexp.MustCompile(ENCAP_REGEXP)
	ip_compiled     = regexp.MustCompile(IP_REGEXP)
	vrf_compiled    = regexp.MustCompile(VRF_REGEXP)
	mtu_compiled    = regexp.MustCompile(MTU_REGEXP)
	aclin_compiled  = regexp.MustCompile(ACLIN_REGEXP)
	aclout_compiled = regexp.MustCompile(ACLOUT_REGEXP)
)



func getIP(s string, d string) (ip_addr, subnet string, subnetRaw netip.Prefix) {

	if strings.Contains(s, "dhcp") {
		return "dhcp", "dhcp", netip.Prefix{}
	}

	if d == "ios" {
		ip_str := ip_compiled.FindStringSubmatch(s)[1]
		mask_str := ip_compiled.FindStringSubmatch(s)[2]
		
		ip := netip.MustParseAddr(ip_str)
		mask, _ := net.IPMask(net.ParseIP(mask_str).To4()).Size()
		prefix := netip.PrefixFrom(ip, mask)

		return prefix.String(), prefix.Masked().String(), prefix.Masked()

	} else if d == "nxos" {
		ip_str := regexp.MustCompile(` {2}ip address (\S+)`).FindStringSubmatch(s)[1]
		prefix := netip.MustParsePrefix(ip_str)
		return ip_str, prefix.Masked().String(), prefix.Masked()
	}
	return
}

// ParseInterfaces func reads config from r, and parses interfaces from it to 'CiscoInterfaceMap' data type.
// Platform type d specifies config origin (IOS or NXOS)

// parse parses config from d.source and populates internal fields "interfaces" and "subnets" 
func(d *Device) parse() error {

	// d.interfaces = CiscoInterfaceMap{}
	var intf_name string

	line_separator := "!"
	line_ident := " "

	if d.platform == NXOS {
		line_separator = ""
		line_ident = "  "
	}

	scanner := bufio.NewScanner(d.source)
	for scanner.Scan() {

		line := strings.TrimRight(scanner.Text(), " ")
		// fmt.Println(line)	// for debug

		if strings.HasPrefix(line, `interface `) { //Enter interface configuration block

			intf_name = intf_compiled.FindStringSubmatch(line)[1]
			d.interfaces[intf_name] = &CiscoInterface{Name: intf_name}

		} else if strings.HasPrefix(line, line_ident) && d.intfAmount() > 0 { //Content inside interface config

			switch {
			case strings.Contains(line, ` description `):
				intf_desc := desc_compiled.FindStringSubmatch(line)[1]
				d.interfaces[intf_name].Description = intf_desc

			case strings.Contains(line, ` encapsulation`):
				encap := encap_compiled.FindStringSubmatch(line)[1]
				d.interfaces[intf_name].Encapsulation = encap

			case strings.Contains(line, `ip address `) || strings.Contains(line, `ipv4 address `):
				ip_cidr, prefix, prefixRaw := getIP(scanner.Text(), d.platform)
				d.interfaces[intf_name].Ip_addr = ip_cidr
				d.interfaces[intf_name].Subnet = prefix
				d.interfaces[intf_name].subnetRaw = prefixRaw

			case strings.Contains(line, ` vrf `):
				vrf := vrf_compiled.FindStringSubmatch(line)[1]
				d.interfaces[intf_name].Vrf = vrf

			case strings.Contains(line, ` mtu `):
				mtu := mtu_compiled.FindStringSubmatch(line)[1]
				d.interfaces[intf_name].Mtu = mtu

			case strings.Contains(line, `access-group `) && strings.HasSuffix(line, ` in`):
				aclin := aclin_compiled.FindStringSubmatch(line)[1]
				d.interfaces[intf_name].ACLin = aclin

			case strings.Contains(line, `access-group `) && strings.HasSuffix(line, ` out`):
				aclout := aclout_compiled.FindStringSubmatch(line)[1]
				d.interfaces[intf_name].ACLout = aclout
			}

		} else if !(line == line_separator || strings.HasPrefix(line, `interface`)) && d.intfAmount() > 0 { //Exit interface configuration block
			break
		}
	}
	if d.intfAmount() == 0 {
		errorLogger.Println("Parsing failed! got 0 interfaces!")
		return ErrParsingFailed
	}
	infoLogger.Printf("parsing finished, got %v interfaces", d.intfAmount())
	return nil
}
