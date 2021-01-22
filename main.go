package main

import (
	"os"

	"github.com/SwingbyProtocol/node-installer/bot"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)
	key := os.Getenv("BOT_TOKEN")
	if key == "" {
		panic("Error: BOT_TOKEN is null")
	}
	bot, err := bot.NewBot(key)
	if err != nil {
		panic(err)
	}
	bot.Start()
	select {}
}
