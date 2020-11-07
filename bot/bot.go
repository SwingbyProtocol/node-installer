package bot

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	hostsFilePath  = "./data/hosts"
	sshKeyFilePath = "./data/ssh_key"
	sshAgtCMD      = "eval $(ssh-agent -s) > /dev/null && ssh-add " + sshKeyFilePath
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
		if !b.validateChat(update.Message.Chat.ID) && b.ID != 0 {
			continue
		}

		if b.mode == "catchIP" {
			ipText := update.Message.Text
			ipAddr := net.ParseIP(ipText)
			if ipAddr == nil {
				text := fmt.Sprintf("IP address should be version 4. Kindly put again")
				b.SendMsg(b.ID, text)
				continue
			}
			generateHostsfile(ipText, "server")
			text := fmt.Sprintf("Your server IP is %s ", ipText)
			b.SendMsg(b.ID, text)
			b.turnOutMode("saveIP")
			text2 := fmt.Sprintf("Please let me know your ssh private key, (only stored into this machine)")
			b.SendMsg(b.ID, text2)
			continue
		}
		if b.mode == "saveIP" {
			sshKey := update.Message.Text
			generateSSHKeyfile(sshKey)
			if verifySSHkey() != nil {
				text := fmt.Sprintf("SSH priv key is not valid. Kindly put again")
				b.SendMsg(b.ID, text)
				continue
			}
			b.turnOutMode("saveSSHKey")
			text2 := fmt.Sprintf("Your server is ready. Please kindly do /setup_infura")
			b.SendMsg(b.ID, text2)
			continue
		}
		if update.Message.Text == "/start" {
			if b.ID == 0 {
				b.ID = update.Message.Chat.ID
			}
			b.SendMsg(b.ID, makeHelloText())
			continue
		}
		if update.Message.Text == "/setup_server" {
			b.SendMsg(b.ID, makeDeployText())
			b.turnOutMode("catchIP")
			continue
		}
		if update.Message.Text == "/setup_infura" {
			command := fmt.Sprintf(
				"%s && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -v -i %s %s --extra-vars '%s'",
				sshAgtCMD,
				hostsFilePath,
				"./playbooks/testnet_infura.yml", "BOT_TOKEN="+b.bot.Token)
			cmd := exec.Command("sh", "-c", command)
			log.Info(cmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Info(err)
				continue
			}
			continue
		}
	}
}

func (b *Bot) validateChat(chatID int64) bool {
	if b.ID == chatID {
		return true
	}
	return false
}

func (b *Bot) turnOutMode(mode string) {
	b.mode = mode
}

func verifySSHkey() error {
	cmd := exec.Command("sh", "-c", sshAgtCMD)
	log.Info(cmd)
	err := cmd.Run()
	return err
}

func (b *Bot) SendMsg(id int64, text string) {
	msg := tgbotapi.NewMessage(id, text)
	msg.ParseMode = "HTML"
	b.bot.Send(msg)
}

func generateHostsfile(nodeIP string, target string) {
	text := fmt.Sprintf("[%s]\n%s", target, nodeIP)
	err := ioutil.WriteFile(hostsFilePath, []byte(text), 0666)
	if err != nil {
		os.Exit(1)
	}
}

func generateSSHKeyfile(key string) {
	text := fmt.Sprintf("%s\n", key)
	err := ioutil.WriteFile(sshKeyFilePath, []byte(text), 0600)
	if err != nil {
		os.Exit(1)
	}
}

func makeHelloText() string {
	text := fmt.Sprintf(`
Hello 😊, This is a deploy bot
Steps is here. 
1. Put /setup_server to setup your server
2. Put /setup_infura to deploy infura into your server
	`)
	return text
}

func makeDeployText() string {
	text := fmt.Sprintf(`
Deploy is starting
Please let me know your server IP address (Only accept Version 4)
	`)
	return text
}
