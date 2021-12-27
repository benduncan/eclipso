package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/benduncan/eclipso/pkg/backend"
)

const eclipso_version = "1.0.1"

func main() {

	var zone_dir = os.Getenv("ZONE_DIR")

	if zone_dir == "" {
		zone_dir = "config/domains/"
	}

	var host = os.Getenv("HOST")

	if host == "" {
		host = "0.0.0.0"
	}

	var port = os.Getenv("PORT")

	if port == "" {
		port = "53"
	}

	log.Printf(`


	┌─┐┌─┐┬  ┬┌─┐┌─┐┌─┐
	├┤ │  │  │├─┘└─┐│ │
	└─┘└─┘┴─┘┴┴  └─┘└─┘	
	High-performance DNS daemon
	v%s

	
	`, eclipso_version)

	err := backend.StartDaemon(zone_dir, host, port)

	if err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}

}
