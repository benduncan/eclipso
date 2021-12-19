package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	SOA      string
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

	var s3retry = os.Getenv("S3_RETRY")

	if s3retry == "" {
		s3retry = "60"
	}

	s3retrysecs, _ := strconv.Atoi(s3retry)

	if strings.HasPrefix(zone_dir, "s3://") {

		go func() {

			for {

				time.Sleep(time.Second * time.Duration(s3retrysecs))

				fmt.Println("In loop to check state")

				sess := session.Must(session.NewSession())

				// Create S3 service client
				svc := s3.New(sess)

				path := strings.Split(zone_dir, "s3://")

				if len(path) == 0 {
					log.Fatal("S3_BUCKET field required")
				}

				// Get the list of items
				resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(path[1])})
				if err != nil {
					log.Fatalf("Unable to list items in bucket %q, %v", path, err)
				}

				for _, item := range resp.Contents {

					if strings.HasSuffix(*item.Key, ".toml") {

						domain := strings.Replace(*item.Key, ".toml", "", 1)

						for i := 0; i < len(HostedZones); i++ {

							if HostedZones[i].Domain.Domain == domain {
								fmt.Println(domain, "> ", HostedZones[i].Domain.Modified, "=>", *item.LastModified)

								if *item.LastModified != HostedZones[i].Domain.Modified {
									fmt.Println("WE HAVE A NEW CONFIG FILE, RELOAD!")

									mu.Lock()
									HostedZones[i], err = ReadZone(fmt.Sprintf("%s/%s", zone_dir, *item.Key), *item.LastModified)
									mu.Unlock()

									if err != nil {
										log.Fatalf("Error %s", err)
									}

								}
							}

						}

						if err != nil {
							log.Fatalf("Unable to download item %q, %v", item, err)
						}

					}

				}

			}

		}()

	} else {

		// Listen to FS events for new/modified files, and reload our state
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
						//myconf, _ := ReadZone(event.Name, event.Modified)
						//fmt.Println(myconf)
						ReadZoneFiles(zone_dir)
						//reloadConf()
					}

					if event.Op&fsnotify.Create == fsnotify.Create {
						log.Println("new file:", event.Name)
						//myconf, _ := ReadZone(event.Name)
						//fmt.Println(myconf)
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

}

func ApplyDefaults(config *Config, lastModified time.Time) {

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

	// Set the lastModified time if specified (e.g from S3 LastModified attribute)
	if lastModified.IsZero() == false {
		config.Domain.Modified = lastModified
	}

	for i := 0; i < len(config.Records); i++ {

		// Set the global TTL if missing
		if config.Records[i].TTL == 0 {
			config.Records[i].TTL = ttl
		}

		// Set as the default record type if missing
		if config.Records[i].Type == 0 {
			config.Records[i].Type = rtype
		}

		// Set default class type
		if config.Records[i].Class == 0 {
			config.Records[i].Class = class
		}

		// Set default MX record preference if undefined
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

	mu.Lock()
	HostedZones = nil

	fmt.Println("Zone dir =>", zone_dir)

	if strings.HasPrefix(zone_dir, "s3://") {

		sess := session.Must(session.NewSession())

		// Create S3 service client
		svc := s3.New(sess)

		path := strings.Split(zone_dir, "s3://")

		if len(path) == 0 {
			log.Fatal("S3_BUCKET field required")
		}

		// Get the list of items
		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(path[1])})
		if err != nil {
			log.Fatalf("Unable to list items in bucket %q, %v", path, err)
		}

		for _, item := range resp.Contents {

			if strings.HasSuffix(*item.Key, ".toml") {
				myconfig, err := ReadZone(fmt.Sprintf("%s/%s", zone_dir, *item.Key), *item.LastModified)

				if err != nil {
					log.Errorf("Error parsing file %s", err)
				}

				HostedZones = append(HostedZones, myconfig)

				if err != nil {
					log.Fatalf("Unable to download item %q, %v", item, err)
				}

			}

		}

	} else {

		files, err := ioutil.ReadDir(zone_dir)

		if err != nil {
			log.Panicf("failed reading directory: %s", err)
		}

		for _, file := range files {

			filename := fmt.Sprintf("%s/%s", zone_dir, file.Name())

			myconfig, err := ReadZone(filename, file.ModTime())

			if err == nil {
				HostedZones = append(HostedZones, myconfig)
			}

		}

		defer mu.Unlock()

	}

	t := time.Now()
	elapsed := t.Sub(start)

	log.Info("Config files read in => ", elapsed)

}

func ReadZone(zone_file string, lastModified time.Time) (myconfig Config, err error) {

	log.Info("Parsing => ", zone_file, lastModified)

	if strings.HasPrefix(zone_file, "s3://") {

		s3path := strings.SplitN(zone_file, "s3://", -1)
		paths := strings.SplitN(s3path[1], "/", 2)

		if len(paths) == 0 {
			return myconfig, errors.New("Path not found in S3")
		}

		sess := session.Must(session.NewSession())

		// Create S3 service client
		//svc := s3.New(sess)

		buff := &aws.WriteAtBuffer{}
		downloader := s3manager.NewDownloader(sess)

		numBytes, _ := downloader.Download(buff,
			&s3.GetObjectInput{
				Bucket: aws.String(paths[0]),
				Key:    aws.String(paths[1]),
			})

		if numBytes > 0 {
			toml.Unmarshal(buff.Bytes(), &myconfig)
			ApplyDefaults(&myconfig, lastModified)

		} else {
			return myconfig, errors.New("Config file empty")
		}

	} else {

		file, err := os.ReadFile(zone_file)

		if err != nil {
			errorMsg := fmt.Sprintf("Error reading %s %s", zone_file, err)
			log.Warn(errorMsg)
			return myconfig, errors.New(errorMsg)
		}

		toml.Unmarshal(file, &myconfig)
		ApplyDefaults(&myconfig, lastModified)

	}

	return

}
