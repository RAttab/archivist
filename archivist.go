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
		Token string              `json:"token"`
		Subs  map[string][]string `json:"subs"`
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
		Fatal("unable to read config '%v': %v", *configPath, err)
	}

	if err := json.Unmarshal(data, &Config); err != nil {
		Fatal("unable to parse config '%v': %v", *configPath, err)
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

	ScrapperStart()
	ApiInit()

	Info("Started")
	notif := make(chan os.Signal, 1)
	signal.Notify(notif, os.Interrupt)
	<-notif

	Info("Exiting")
}

func Fatal(fmt string, args ...interface{}) {
	log.Fatalf("<FATAL> "+fmt, args...)
}

func Warning(fmt string, args ...interface{}) {
	log.Printf("<WARN> "+fmt, args...)
}

func Info(fmt string, args ...interface{}) {
	log.Printf("<INFO> "+fmt, args...)
}

func Debug(fmt string, args ...interface{}) {
	log.Printf("<DEBUG> "+fmt, args...)
}
