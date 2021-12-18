package main

import (
	"fmt"
	"os"

	"github.com/benduncan/eclipso/pkg/backend"
	"github.com/benduncan/eclipso/pkg/config"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

func main() {

	var zone_dir = os.Getenv("ZONE_DIR")

	if zone_dir == "" {
		zone_dir = "config/domains/"
	}

	config.ReadZoneFiles(zone_dir)

	backend.DomainDB = &config.HostedZones

	go config.MonitorConfig(zone_dir)

	var host = os.Getenv("HOST")

	if host == "" {
		host = "0.0.0.0"
	}

	var port = os.Getenv("PORT")

	if port == "" {
		port = "53"
	}

	srv := &dns.Server{Addr: fmt.Sprintf("%s:%s", host, port), Net: "udp"}
	srv.Handler = &backend.Handler{}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
