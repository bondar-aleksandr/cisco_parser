package cisco_parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"reflect"
	"testing"
	"fmt"
)

var testDataDir = "./test_data"

func deviceFromFile(filename string) *Device {
	jsonFile, err := os.ReadFile(filename)
	if err != nil {
		infoLogger.Fatalf("Cannot open file %s", filename)
	}

	device := &Device{}
	err = json.Unmarshal(jsonFile, device)
	if err != nil {
		infoLogger.Fatalf("Cannot deserialize file %s into JSON", filename)
	}
	return device
}

func Test_Parsing_Subnets(t *testing.T) {

	ios_ifile_router := filepath.Join(testDataDir, "INET-R01.txt")
	ios_ifiile_switch := filepath.Join(testDataDir, "run.txt")
	ios_ifile_routerXR := filepath.Join(testDataDir, "ASR-P.txt")
	nxos_ifile := filepath.Join(testDataDir, "dc0-n9k-d_23.08.txt")

	ios_device_router := deviceFromFile(fileExtReplace(ios_ifile_router, "json"))
	ios_device_switch := deviceFromFile(fileExtReplace(ios_ifiile_switch, "json"))
	ios_device_routerXR := deviceFromFile(fileExtReplace(ios_ifile_routerXR, "json"))
	nxos_device := deviceFromFile(fileExtReplace(nxos_ifile, "json"))

	configs := []struct {
		name     string
		ifile    string
		platform string
		expected *Device
	}{
		{name: "ios-router", ifile: ios_ifile_router, platform: "ios", expected: ios_device_router},
		{name: "ios-L3switch", ifile: ios_ifiile_switch, platform: "ios", expected: ios_device_switch},
		{name: "ios-XR", ifile: ios_ifile_routerXR, platform: "ios", expected: ios_device_routerXR},
		{name: "NXOS", ifile: nxos_ifile, platform: "nxos", expected: nxos_device},
	}

	for _, v := range configs {

		ifile := v.ifile
		platform := v.platform
		target_device := v.expected
		f, err := os.Open(ifile)
		if err != nil {
			t.Errorf("Cannot open configuration file %s because of %q", ifile, err)
		}
		device, _ := NewDevice(f, platform)
		if err = device.parse(); err != nil {
			t.Errorf("can't parse config: %s", err)
		}
		eq := reflect.DeepEqual(device.Interfaces, target_device.Interfaces)
		if !eq {
			t.Errorf("%s: parsed config doesn't correspond target value", v.name)
		}
	}
}

func fileExtReplace(f string, ex string) string {
	bareName := strings.TrimSuffix(f, filepath.Ext(f))
	return fmt.Sprintf("%s.%s", bareName, ex)
}