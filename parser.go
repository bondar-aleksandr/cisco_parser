package cisco_parser

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strings"
)

var ErrParsigFailed = errors.New("no interfaces found in config")


const (
	hostname_regexp = `hostname (\S+)`
	intf_regexp   = `^interface (\S+)`
	desc_regexp   = ` {1,2}description (.*)$`
	encap_regexp  = ` {1,2}encapsulation (.+)`
	ip_regexp     = ` {1,2}ip(?:v4)? address (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(?: secondary)?`
	ip_regexp_nxos = ` {2}ip address (\S+)`
	vrf_regexp    = ` {1,2}vrf(?: forwarding| member)? (\S+)`
	mtu_regexp    = ` {1,2}(?:ip )?mtu (\S+)`
	aclin_regexp  = ` {1,2}access-group (\S+) in`
	aclout_regexp = ` {1,2}access-group (\S+) out`
)

var (
	hostname_compiled = regexp.MustCompile(hostname_regexp)
	intf_compiled   = regexp.MustCompile(intf_regexp)
	desc_compiled   = regexp.MustCompile(desc_regexp)
	encap_compiled  = regexp.MustCompile(encap_regexp)
	ip_compiled     = regexp.MustCompile(ip_regexp)
	ip_compiled_nxos     = regexp.MustCompile(ip_regexp_nxos)
	vrf_compiled    = regexp.MustCompile(vrf_regexp)
	mtu_compiled    = regexp.MustCompile(mtu_regexp)
	aclin_compiled  = regexp.MustCompile(aclin_regexp)
	aclout_compiled = regexp.MustCompile(aclout_regexp)
)


// getIP parses ip address and mask into strings and into netip.Prefix. Returns the values
// and error if any occured duing parsing.
func getIP(s string, d string) (ip_addr, subnet string, subnetRaw netip.Prefix, err error) {
	
	emptyPrefix := netip.Prefix{}

	if strings.Contains(s, "dhcp") {
		return "dhcp", "dhcp", emptyPrefix, nil
	}

	if d == "ios" {
		ip_str := ip_compiled.FindStringSubmatch(s)[1]
		mask_str := ip_compiled.FindStringSubmatch(s)[2]
		
		ip, err := netip.ParseAddr(ip_str)
		if err != nil {
			return "", "", emptyPrefix, fmt.Errorf("can't parse IP: %w", err)
		}
		mask, _ := net.IPMask(net.ParseIP(mask_str).To4()).Size()
		prefix := netip.PrefixFrom(ip, mask)

		return prefix.String(), prefix.Masked().String(), prefix.Masked(), nil

	} else if d == "nxos" {
		ip_str := ip_compiled_nxos.FindStringSubmatch(s)[1]
		prefix, err := netip.ParsePrefix(ip_str)
		if err != nil {
			return "", "", emptyPrefix, fmt.Errorf("can't parse IP: %w", err)
		}
		return ip_str, prefix.Masked().String(), prefix.Masked(), nil
	}
	return
}

// parse parses config from d.source and populates internal fields d.interfaces and d.subnets
func(d *Device) parse() error {

	var intf_name string
	var intf *CiscoInterface
	var intfVrf *intfVrf

	line_separator := "!"
	line_ident := " "

	if d.platform == nxos {
		line_separator = ""
		line_ident = "  "
	}

	scanner := bufio.NewScanner(d.source)
	for scanner.Scan() {

		line := strings.TrimRight(scanner.Text(), " ")
		// fmt.Println(line)	// for debug

		if strings.HasPrefix(line, `hostname `) {	// parse hostname
			hostname := hostname_compiled.FindStringSubmatch(line)[1]
			d.Hostname = hostname

		} else if strings.HasPrefix(line, `interface `) { //Enter interface configuration block
			
			intf_name = intf_compiled.FindStringSubmatch(line)[1]
			intf = newCiscoInterface(intf_name)
			if err := d.addInterface(intf); err != nil {
				return fmt.Errorf("can't parse: %w", err)
			}

		} else if strings.HasPrefix(line, line_ident) && d.intfAmount() > 0 { //Content inside interface config

			switch {
			case strings.Contains(line, ` description `):
				intf_desc := desc_compiled.FindStringSubmatch(line)[1]
				intf.Description = intf_desc

			case strings.Contains(line, ` encapsulation`):
				encap := encap_compiled.FindStringSubmatch(line)[1]
				intf.Encapsulation = encap

			case strings.Contains(line, ` vrf `):
				vrf := vrf_compiled.FindStringSubmatch(line)[1]
				intf.Vrf = vrf

			case strings.Contains(line, `ip address `) || strings.Contains(line, `ipv4 address `):
				ip_cidr, prefix, prefixRaw, err := getIP(scanner.Text(), d.platform)
				if err != nil {
					ip_cidr, prefix = "FAILED TO PARSE", "FAILED TO PARSE"
					warnLogger.Println("failed to parse ip:", err)
				}
				intf.Ip_addr = ip_cidr
				intf.Subnet = prefix

				intfVrf = newInterfaceVrf(intf.Name)
				intfVrf.addVrf(intf.Vrf)
				d.addSubnet(prefixRaw, intfVrf)

			case strings.Contains(line, ` mtu `):
				mtu := mtu_compiled.FindStringSubmatch(line)[1]
				intf.Mtu = mtu

			case strings.Contains(line, `access-group `) && strings.HasSuffix(line, ` in`):
				aclin := aclin_compiled.FindStringSubmatch(line)[1]
				intf.ACLin = aclin

			case strings.Contains(line, `access-group `) && strings.HasSuffix(line, ` out`):
				aclout := aclout_compiled.FindStringSubmatch(line)[1]
				intf.ACLout = aclout
			}

		} else if !(line == line_separator || strings.HasPrefix(line, `interface`)) && d.intfAmount() > 0 { //Exit interface configuration block
			break
		}
	}
	if d.intfAmount() == 0 {
		errorLogger.Println("Parsing failed! got 0 interfaces!")
		return ErrParsigFailed
	}
	d.parsed = true
	infoLogger.Printf("parsing finished, got %v interfaces", d.intfAmount())
	return nil
}