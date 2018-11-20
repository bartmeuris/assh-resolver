package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/jackpal/gateway"
	yaml "gopkg.in/yaml.v2"
)

// Version is set by the build process to contain the version of the application
var Version string

// Build is set by the build process to contain the git hash of the compiled program
var Build string

// Debug is a string set by the build process to indicate debug builds
var Debug string
var debugBool bool

// EnvVarName is the name of the environment variable that is consulted to locate the configuration yaml file
const EnvVarName = "ASSH_RESOLVECFG"

// ConfigFileName is the filename that is searched for in in the ~/.ssh folder
const ConfigFileName = "locations.yml"

// HostSeparator is the separator used to split the incoming host argument locations
const HostSeparator = "|"

// LocationSeparator is the separator used to split the location/host
const LocationSeparator = ";"

const usageText = `
`

// Location contains the structure used to parse the YAML config file
type Location struct {
	Gateway string
	Short   string
	Name    string `yaml:"omitempty"`
}

func (l *Location) String() string {
	return fmt.Sprintf("%s (%s) with gateway %s", l.Name, l.Short, l.Gateway)
}
func checkError(err error, format string, a ...interface{}) {
	if err == nil {
		return
	}
	if format != "" {
		fmt.Fprintf(os.Stderr, format, a...)
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n\n", err)
	flag.Usage()
	os.Exit(1)
}

func debug(format string, a ...interface{}) {
	if debugBool {
		fmt.Fprintf(os.Stderr, format, a...)
		if !strings.Contains(format, "\n") {
			fmt.Fprint(os.Stderr, "\n")
		}
	}
}

func findLocation(configfile string) (*Location, error) {
	var locs map[string]*Location
	defReturn := &Location{Short: "default", Name: "", Gateway: ""}

	cdata, err := ioutil.ReadFile(configfile)
	if err != nil {
		debug("  Could not read config file '%s': %s", configfile, err)
		return defReturn, err
	}

	err = yaml.Unmarshal(cdata, &locs)
	if err != nil {
		debug("  Could not parse config file '%s': %s", configfile, err)
		return defReturn, err
	}

	gw, err := gateway.DiscoverGateway()
	if err != nil {
		debug("  Could not find default gateway: %s", err)
		return defReturn, err
	}
	gws := fmt.Sprintf("%s", gw)
	debug("  Detected Gateway: %s\n", gws)

	for s := range locs {
		if locs[s].Name == "" {
			// Don't set name if it was overridden in the config file
			locs[s].Name = s
		}
		if locs[s].Gateway == gws {
			debug("  Found matching location: %s", locs[s].Name)
			return locs[s], nil
		}
		// No gateway defined -> this is the default entry
		if locs[s].Gateway == "" {
			debug("  Encountered location without gateway, setting as default location: %s (%s)", s, locs[s].Short)
			defReturn = locs[s]
		}
	}

	return defReturn, nil
}

func fileReadable(name string) bool {
	if f, err := os.Open(name); err != nil {
		return false
	} else if err = f.Close(); err != nil {
		return false
	}
	return true
}

func defaultConfigFile() string {
	// Only try to open the config file in current directory in debug builds
	if debugBool && fileReadable(ConfigFileName) {
		return ConfigFileName
	}

	val, ok := os.LookupEnv(EnvVarName)
	if ok && fileReadable(val) {
		return val
	}
	if usr, err := user.Current(); err == nil {
		fn := fmt.Sprintf("%s%c%s%c%s", usr.HomeDir, os.PathSeparator, ".ssh", os.PathSeparator, ConfigFileName)
		if fileReadable(fn) {
			return fn
		}
	}

	return ""
}

func getLocIP(loc Location, hoststring string) (string, error) {
	ips := strings.Split(hoststring, HostSeparator)

	defHost := ""
	host := ""
	for s := range ips {
		//fmt.Printf("s: %s\n", ips[s])
		if cpos := strings.Index(ips[s], LocationSeparator); cpos == -1 {
			// This is an entry without a name, this is the default host
			defHost = ips[s]
			//fmt.Printf("Default host: %s\n", defHost)
		} else {
			hLoc := ips[s][:cpos]
			hHost := ips[s][cpos+1:]
			if defHost == "" {
				// Set the first entry as default host
				defHost = hHost
			}
			if hLoc == loc.Name || hLoc == loc.Short {
				//fmt.Printf("Found match: %s: %s\n", hLoc, hHost)
				if host != "" {
					return "", fmt.Errorf("ERROR: Multiple matching hosts found")
				}
				host = hHost
			}
			//fmt.Printf("Location: %s / Host: %s\n", hLoc, hHost)
		}
	}
	if host == "" {
		//fmt.Printf("Host empty, fallback to default %s\n", defHost)
		host = defHost
	}
	return host, nil
}

func main() {
	var err error
	debugBool, err = strconv.ParseBool(Debug)

	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		if debugBool {
			fmt.Fprintf(flag.CommandLine.Output(), "Version %s (git:%s) (DEBUG BUILD)", Version, Build)
		} else {
			fmt.Fprintf(flag.CommandLine.Output(), "Version %s (git:%s)", Version, Build)
		}
		fmt.Fprint(flag.CommandLine.Output(), usageText)
		flag.PrintDefaults()
	}

	configfile := flag.String("configfile", "", "path to the yaml configuration file")
	flag.BoolVar(&debugBool, "debug", debugBool, "Output debug info to stderr")
	flag.Parse()

	if *configfile == "" {
		*configfile = defaultConfigFile()
	}
	if *configfile == "" || !fileReadable(*configfile) {
		checkError(fmt.Errorf("Could not read configuration file '%s'", *configfile), "")
	}
	debug("Using config file '%s'", *configfile)
	location, err := findLocation(*configfile)
	if location == nil {
		checkError(err, "Could not find location")
	} else if err != nil {
		debug("Warning - problem when detecting location, got: %s with error: %s", location, err)
	} else {
		debug("Detected location: %s", location)
	}

	if len(flag.Args()) != 1 {
		checkError(fmt.Errorf("Expected 1 argument, got %d", len(flag.Args())), "")
	}

	host, err := getLocIP(*location, flag.Args()[0])
	checkError(err, "")
	fmt.Printf("%s\n", host)
}
