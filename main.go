package main

import (
	"github.com/jackpal/gateway"
	"gopkg.in/yaml.v2"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Location struct {
	Gateway string
	Short string
	Name string `yaml:"omitempty"`
}

func checkError(err error, format string, a ...interface{}) {
	if err == nil {
		return
	}
	if format != ""{
		fmt.Printf(format, a)
	}
	fmt.Printf("Error: %s\n", err)
	os.Exit(1)
}

func findLocation(configfile string) Location {
	var locs map[string]*Location

	cdata, err := ioutil.ReadFile(configfile)
	checkError(err, "Could not read %s file", configfile)

	err = yaml.Unmarshal(cdata, &locs)
	checkError(err, "Configuration file in wrong format")

	gw, err := gateway.DiscoverGateway()
	checkError(err, "Could not get default gateway")

	gws := fmt.Sprintf("%s", gw)
	def_return := Location{Short:"default", Name:"", Gateway:""}
	for s := range locs {
		if locs[s].Gateway == gws {
			locs[s].Name = s
			return *locs[s]
		}
		// No gateway defined -> this is the default entry
		if locs[s].Gateway == "" {
			def_return = *locs[s]
		}
	}

	return def_return
}

func main() {
	location := findLocation("locations.yml")
	//fmt.Printf("Location: %v\n", location)
	
	if len(os.Args) < 2 {
		checkError(fmt.Errorf("Expected 1 argument, got %d", len(os.Args)-1), "")
	}
	ips := strings.Split(os.Args[1], "|")
	
	def_host := ""
	host := ""
	for s := range ips {
		//fmt.Printf("s: %s\n", ips[s])
		if cpos := strings.Index(ips[s], ";"); cpos == -1 {
			// This is an entry without a name, this is the default host
			def_host = ips[s]
			//fmt.Printf("Default host: %s\n", def_host)
		} else {
			h_loc := ips[s][:cpos]
			h_host := ips[s][cpos+1:]
			if def_host == "" {
				// Set the first entry as default host
				def_host = h_host
			}
			if h_loc == location.Name || h_loc == location.Short {
				//fmt.Printf("Found match: %s: %s\n", h_loc, h_host)
				if host != "" {
					checkError(fmt.Errorf("ERROR: Multiple matching hosts found"), "")
				}
				host = h_host
			}
			//fmt.Printf("Location: %s / Host: %s\n", h_loc, h_host)
		}
	}
	if host == "" {
		//fmt.Printf("Host empty, fallback to default %s\n", def_host)
		host = def_host
	}
	fmt.Printf("%s\n", host)
}
