package main

import (
	"github.com/jackpal/gateway"
	"gopkg.in/yaml.v2"
	"fmt"
	"io/ioutil"
	"os"
)

type Location struct {
	Gateway string
	Short string
	Name string `yaml:"omitempty"`
}

func main() {
	var s map[string]Location
	
	cdata, err := ioutil.ReadFile("/tmp/dat")
	yaml.Unmarshal(cdata, &s)

	ip, err := gateway.DiscoverGateway()
	if err  != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Gateway: %s\n", ip)
}
