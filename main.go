package main

import (
	"os"

	"github.com/SwingbyProtocol/node-installer/services/bot"
)

func main() {
	key := os.Getenv("BOT_TOKEN")
	bot, err := bot.NewBot(key)
	if err != nil {
		panic(err)
	}
	bot.Start()
	select {}
}
