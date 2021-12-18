package main

import (
	"fmt"
	"log"

	"github.com/benduncan/eclipso/pkg/config"
	"github.com/fsnotify/fsnotify"
)

var hostedzones []config.Config

func main() {

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
					myconf, _ := config.ReadZone(event.Name)
					fmt.Println(myconf)
					reloadConf()
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("new file:", event.Name)
					myconf, _ := config.ReadZone(event.Name)
					fmt.Println(myconf)
					reloadConf()

				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("remove file:", event.Name)
					reloadConf()
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

	err = watcher.Add("config/tmp")
	if err != nil {
		log.Fatal(err)
	}

	<-done

}

func reloadConf() {
	config.ReadZoneFiles("config/tmp")
}
