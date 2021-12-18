package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	mu       sync.Mutex
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
	Type  uint16
	Class uint16
}

type Records struct {
	Domain     string
	TTL        uint32
	Type       uint16
	Class      uint16
	Preference uint16
	Address    string
}

var HostedZones []Config
var mu sync.Mutex

func init() {
	_, logignore := os.LookupEnv("ECLIPSO_LOG_IGNORE")

	if logignore {
		log.SetLevel(log.FatalLevel)
	}
}

func MonitorConfig(zone_dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					myconf, _ := ReadZone(event.Name)
					fmt.Println(myconf)
					ReadZoneFiles(zone_dir)
					//reloadConf()
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("new file:", event.Name)
					myconf, _ := ReadZone(event.Name)
					fmt.Println(myconf)
					ReadZoneFiles(zone_dir)
					//reloadConf()

				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("remove file:", event.Name)
					ReadZoneFiles(zone_dir)
					//reloadConf()
					//myconf := config.ReadZone(event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(zone_dir)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func ApplyDefaults(config *Config) {

	var ttl uint32
	var rtype uint16
	var class uint16

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

		if config.Records[i].Type == 15 && config.Records[i].Preference == 0 {
			config.Records[i].Preference = 10
		}

		// Next append the root domain to the record
		config.Records[i].Domain = fmt.Sprintf("%s%s.", config.Records[i].Domain, config.Domain.Domain)

		// Check record size, 255 bytes max
		rsize := len(config.Records[i].Address)

		if rsize > 255 {
			config.Records[i].Address = config.Records[i].Address[:255]
			log.Warn(config.Records[i].Domain, " => Record size too large, 255 byte limit, truncated.")
		}

	}

}

func ReadZoneFiles(zone_dir string) {

	start := time.Now()

	files, err := ioutil.ReadDir(zone_dir)

	if err != nil {
		log.Panicf("failed reading directory: %s", err)
	}

	mu.Lock()
	HostedZones = nil

	for _, file := range files {

		filename := fmt.Sprintf("%s/%s", zone_dir, file.Name())

		myconfig, err := ReadZone(filename)

		if err == nil {
			HostedZones = append(HostedZones, myconfig)
		}

	}

	t := time.Now()
	elapsed := t.Sub(start)

	log.Info("Config files read in => ", elapsed)

	defer mu.Unlock()
}

func ReadZone(zone_file string) (myconfig Config, err error) {

	log.Info("Parsing => ", zone_file)
	file, err := os.ReadFile(zone_file)

	if err != nil {
		errorMsg := fmt.Sprintf("Error reading %s %s", zone_file, err)
		log.Warn(errorMsg)
		return myconfig, errors.New(errorMsg)
	}

	toml.Unmarshal(file, &myconfig)
	ApplyDefaults(&myconfig)

	return

}
