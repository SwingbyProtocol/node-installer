package bot

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	mu  *sync.RWMutex
	bot *tgbotapi.BotAPI
	ID  int64
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
	}
}

func (b *Bot) SendMsg(text string) {
	msg := tgbotapi.NewMessage(b.ID, text)
	msg.ParseMode = "HTML"
	b.bot.Send(msg)
}
