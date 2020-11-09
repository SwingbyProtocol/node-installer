package main

import (
	"os"

	"github.com/SwingbyProtocol/node-installer/bot"
)

func main() {
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
