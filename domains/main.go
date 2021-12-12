package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Version  float32
	Domain   Domain
	Defaults Defaults
	Records  []Records
}

type Domain struct {
	Domain   string
	Created  time.Time
	Modified time.Time
	Verified bool
	Active   bool
	OwnerID  uint32
}

type Defaults struct {
	TTL   uint32
	Type  uint8
	Class uint8
}

type Records struct {
	Domain  string
	TTL     uint32
	Type    uint8
	Class   uint8
	Address string
}

func main() {

	file, err := os.ReadFile("phasegrid.net.toml")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config := Config{}
	toml.Unmarshal(file, &config)
	applyDefaults(&config)

	fmt.Println(config)

}

func applyDefaults(config *Config) {

	var ttl uint32
	var rtype uint8
	var class uint8

	if config.Defaults.TTL > 0 {
		ttl = config.Defaults.TTL
	} else {
		ttl = 3600
	}

	if config.Defaults.Type > 0 {
		rtype = config.Defaults.Type
	} else {
		rtype = 1
	}

	if config.Defaults.Class > 0 {
		class = config.Defaults.Class
	} else {
		class = 1
	}

	for i := 0; i < len(config.Records); i++ {

		if config.Records[i].TTL == 0 {
			config.Records[i].TTL = ttl
		}

		if config.Records[i].Type == 0 {
			config.Records[i].Type = rtype
		}

		if config.Records[i].Class == 0 {
			config.Records[i].Class = class
		}

	}

}
