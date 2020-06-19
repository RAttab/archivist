package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

var configPath = flag.String("config", "./archivist.json", "JSON config")
var Config struct {
	Token        string `json:"token"`
	DatabasePath string `json:"db"`
	Guild        string `json:"guild"`
	Channel      string `json:"channel"`
	Bind         string `json:"bind"`
}

func ConfigLoad() {
	flag.Parse()

	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("unable to read config '%v': %v", *configPath, err)
	}

	if err := json.Unmarshal(data, &Config); err != nil {
		log.Fatalf("unable to parse config '%v': %v", *configPath, err)
	}
}

func main() {
	ConfigLoad()

	DatabaseOpen(Config.DatabasePath)
	defer DatabaseClose()

	DiscordConnect()
	defer DiscordClose()

	ScrapperStart(Config.Guild, Config.Channel, DatabaseIt(Config.Channel))
	defer ScrapperStop()

	ApiInit()

	log.Printf("started...")
	notif := make(chan os.Signal, 1)
	signal.Notify(notif, os.Interrupt)
	<-notif
}
