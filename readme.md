# Cisco parser package

`cisco_parser`  package provides cisco IOS/IOS-XE/IOS-XR/NXOS config parsing capability. It parses interfaces part of configuration. As input, it uses configuration data (as io.Reader). The output is csv or json formatted data (as io.Writer).

## Usage
We need to specify the following info:
- io.Reader, to parse config data from (file, buffer, os.Stdin, whatever)
- source device platform, from where config-file is taken. Allowed values are `ios`, `nxos`
- io.Writer, to write structured data to
- output data format. Allowed values are `csv`, `json`

Code example:
```go

package main
import 	(
    "github.com/bondar-aleksandr/cisco_parser"
    "os"
)

func main() {
    // source config file
    inputFile, err := os.Open("config/file/name")
    if err != nil {
        ...
    }
    defer inputFile.Close()

    // source device platform
    platform = "ios"

    // create device model
    device, err := cisco_parser.NewDevice(inputFile, platform)
    if err != nil {
        ...
    }

    // prepare output file
    outputFile, err := os.Create("path/to/output")
    if err != nil {
        ...
    }
    defer outputFile.Close()

    //prepare serializer
    format = "csv"
    serializer, err := cisco_parser.NewSerializer(outputFile, device, format)
    if err != nil {
        ...
	}
    if err = serializer.Serialize(); err != nil {
        ...
    }
}

```

