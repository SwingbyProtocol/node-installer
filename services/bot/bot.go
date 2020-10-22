package bot

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	mu   *sync.RWMutex
	bot  *tgbotapi.BotAPI
	ID   int64
	mode string
}

func NewBot(token string) (*Bot, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot := &Bot{
		mu:  new(sync.RWMutex),
		bot: b,
		ID:  0,
	}
	return bot, nil
}

func (b *Bot) Start() {
	b.bot.Debug = false
	log.Printf("Authorized on account %s\n", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}
	for update := range updates {
		log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

		if b.mode == "catchIP" {
			ip := update.Message.Text
			generateHostsfile(ip, "server")
			text := fmt.Sprintf("Your server IP is %s ", ip)
			b.SendMsg(b.ID, text)
			b.turnOutMode("saveIP")
			text2 := fmt.Sprintf("Please let me know your ssh private key, (only stored into this machine)")
			b.SendMsg(b.ID, text2)
		}
		if update.Message.Text == "/start" {
			if b.ID == 0 {
				b.ID = update.Message.Chat.ID
			}
			b.SendMsg(b.ID, makeHelloText())
		}
		if update.Message.Text == "/deploy" {
			b.SendMsg(b.ID, makeDeployText())
			b.turnOutMode("catchIP")
		}
	}
}

func (b *Bot) turnOutMode(mode string) {
	b.mode = mode
}

func (b *Bot) SendMsg(id int64, text string) {
	msg := tgbotapi.NewMessage(id, text)
	msg.ParseMode = "HTML"
	b.bot.Send(msg)
}

func generateHostsfile(nodeIP string, target string) {
	text := fmt.Sprintf(`[%s]
%s
	`, target, nodeIP)
	err := ioutil.WriteFile("./configs/hosts", []byte(text), 0666)
	if err != nil {
		os.Exit(1)
	}
}

func makeHelloText() string {
	text := fmt.Sprintf(`
Hello, This is a deploy bot
Steps is here.
1. Put /deploy to start deploying
2. put your ssh key to here
3  .test %s
	`, "base ")
	return text
}

func makeDeployText() string {
	text := fmt.Sprintf(`
Deploy is starting
Please let me know your server IP address (v4)
	`)
	return text
}
