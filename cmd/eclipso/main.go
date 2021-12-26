package main

import (
	"log"
	"os"

	"github.com/benduncan/eclipso/pkg/backend"
)

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

	err := backend.StartDaemon(zone_dir, host, port)

	if err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}

}
