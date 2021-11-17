package main

import (
	"os"
	"os/signal"
	"syscall"

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
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGSTOP)
	<-c
	_, err = bot.SendMsg(bot.ID, bot.BotDownMessage(), false, false)
	if err != nil {
		log.Error(err)
	}
}
