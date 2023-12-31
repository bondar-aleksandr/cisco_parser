package cisco_parser

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type CiscoInterface struct {
	Name        string
	Description string
	Encapsulation string
	Ip_addr     string
	Subnet      string
	Vrf         string
	Mtu         string
	ACLin       string
	ACLout      string
}

func (c CiscoInterface) ToSlice() []string {
	return []string{c.Name, c.Description, c.Encapsulation, c.Ip_addr, c.Subnet, c.Vrf, c.Mtu, c.ACLin, c.ACLout}
}

type CiscoInterfaceMap map[string]*CiscoInterface

func (c CiscoInterfaceMap) GetSortedKeys() []string {
	keys := make([]string, 0)
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (c CiscoInterfaceMap) GetFields() []string {
	fields := reflect.VisibleFields(reflect.TypeOf(CiscoInterface{}))
	result := []string{}
	for _, v := range fields {
		result = append(result, v.Name)
	}
	return result
}

func (c CiscoInterfaceMap) ToJSON(w io.Writer) { // For testing purpose, to get structured data to deserialize from
	json_data, _ := json.MarshalIndent(c, "", "  ")
	_, err := w.Write(json_data)
	if err != nil {
		log.Error("Unable to write json data because of:", err)
	}
	log.Infof("Writing JSON data done")
}

func (c CiscoInterfaceMap) ToCSV(w io.Writer) {
	cw := csv.NewWriter(w)
	headers := c.GetFields()
	cw.Write(headers)

	for _, v := range c.GetSortedKeys() {
		line := c[v].ToSlice()
		cw.Write(line)
	}
	cw.Flush()
	log.Info("Writing CSV data done")
}

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

func FileExtReplace(f string, ex string) string {
	bareName := strings.TrimSuffix(f, filepath.Ext(f))
	return fmt.Sprintf("%s.%s", bareName, ex)
}

func getIP(s string, d string) (ip_addr, subnet string) {

	if strings.Contains(s, "dhcp") {
		return "dhcp", "dhcp"
	}

	if d == "ios" {

		ip_str := ip_compiled.FindStringSubmatch(s)[1]
		mask_str := ip_compiled.FindStringSubmatch(s)[2]

		ip := net.ParseIP(ip_str).To4()
		mask := net.IPMask(net.ParseIP(mask_str).To4())
		mask_cidr, _ := mask.Size()
		net_addr := ip.Mask(mask)
		ip_cidr := fmt.Sprintf("%s/%v", ip.String(), mask_cidr)
		prefix := fmt.Sprintf("%s/%v", net_addr.String(), mask_cidr)

		return ip_cidr, prefix

	} else if d == "nxos" {
		ip_str := regexp.MustCompile(` {2}ip address (\S+)`).FindStringSubmatch(s)[1]
		_, prefix, _ := net.ParseCIDR(ip_str)
		return ip_str, prefix.String()
	}
	return
}

// ParseInterfaces func reads config from r, and parses interfaces from it to 'CiscoInterfaceMap' data type.
// Platform type d specifies config origin (IOS or NXOS)
func ParseInterfaces(r io.Reader, d string) (CiscoInterfaceMap, error) {

	interfaces := CiscoInterfaceMap{}
	var intf_name string

	line_separator := "!"
	line_ident := " "

	if d == "nxos" {
		line_separator = ""
		line_ident = "  "
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {

		line := strings.TrimRight(scanner.Text(), " ")
		// fmt.Println(line)	// for debug

		if strings.HasPrefix(line, `interface `) { //Enter interface configuration block

			intf_name = intf_compiled.FindStringSubmatch(line)[1]
			interfaces[intf_name] = &CiscoInterface{Name: intf_name}

		} else if strings.HasPrefix(line, line_ident) && len(interfaces) > 0 { //Content inside interface config

			switch {
			case strings.Contains(line, ` description `):
				intf_desc := desc_compiled.FindStringSubmatch(line)[1]
				interfaces[intf_name].Description = intf_desc

			case strings.Contains(line, ` encapsulation`):
				encap := encap_compiled.FindStringSubmatch(line)[1]
				interfaces[intf_name].Encapsulation = encap

			case strings.Contains(line, `ip address `) || strings.Contains(line, `ipv4 address `):
				ip_cidr, prefix := getIP(scanner.Text(), d)
				interfaces[intf_name].Ip_addr = ip_cidr
				interfaces[intf_name].Subnet = prefix

			case strings.Contains(line, ` vrf `):
				vrf := vrf_compiled.FindStringSubmatch(line)[1]
				interfaces[intf_name].Vrf = vrf

			case strings.Contains(line, ` mtu `):
				mtu := mtu_compiled.FindStringSubmatch(line)[1]
				interfaces[intf_name].Mtu = mtu

			case strings.Contains(line, `access-group `) && strings.HasSuffix(line, ` in`):
				aclin := aclin_compiled.FindStringSubmatch(line)[1]
				interfaces[intf_name].ACLin = aclin

			case strings.Contains(line, `access-group `) && strings.HasSuffix(line, ` out`):
				aclout := aclout_compiled.FindStringSubmatch(line)[1]
				interfaces[intf_name].ACLout = aclout
			}

		} else if !(line == line_separator || strings.HasPrefix(line, `interface`)) && len(interfaces) > 0 { //Exit interface configuration block
			break
		}
	}
	if len(interfaces) == 0 {
		err := errors.New("parsing failed")
		log.Error("Parsing failed! got 0 interfaces!")
		return interfaces, err
	}
	log.Infof("parsing finished, got %v interfaces", len(interfaces))
	return interfaces, nil
}
