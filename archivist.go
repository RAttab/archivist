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
	Database string `json:"db"`

	Discord struct {
		Token   string `json:"token"`
		Guild   string `json:"guild"`
		Channel string `json:"channel"`
	} `json:"discord"`

	Http struct {
		Assets string `json:"assets"`

		Bind    string `json:"bind"`
		BindTls string `json:"bind_tls"`

		TlsCert string `json:"tls_cert"`
		TlsKey  string `json:"tls_key"`
	} `json:"http"`
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

	if env, ok := os.LookupEnv("DISCORD_TOKEN"); ok {
		Config.Discord.Token = env
	}
}

func main() {
	ConfigLoad()

	DatabaseOpen(Config.Database)
	defer DatabaseClose()

	DiscordConnect()
	defer DiscordClose()

	ScrapperStart(Config.Discord.Guild, Config.Discord.Channel, DatabaseIt(Config.Discord.Channel))
	defer ScrapperStop()

	ApiInit()

	log.Printf("started...")
	notif := make(chan os.Signal, 1)
	signal.Notify(notif, os.Interrupt)
	<-notif
}
