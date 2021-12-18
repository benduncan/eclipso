package main

import (
	"fmt"
	"os"

	"github.com/benduncan/eclipso/pkg/config"
	"github.com/pelletier/go-toml/v2"
)

func main() {

	file, err := os.ReadFile("domains/phasegrid.net.toml")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	myconfig := config.Config{}
	toml.Unmarshal(file, &myconfig)
	config.ApplyDefaults(&myconfig)

	fmt.Println(myconfig)

}
