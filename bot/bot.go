package bot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	ansibler "github.com/apenella/go-ansible"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	hostsFilePath  = "./data/hosts"
	sshKeyFilePath = "./data/ssh_key"
	sshAgtCMD      = "ssh-add " + sshKeyFilePath
)

type Bot struct {
	mu       *sync.RWMutex
	bot      *tgbotapi.BotAPI
	ID       int64
	mode     string
	nodeIP   string
	sshKey   string
	isRemote bool
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
	ansibler.AnsibleAvoidHostKeyChecking()
	b.bot.Debug = false
	log.Printf("Authorized on account %s\n", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	// load host key file
	b.loadBotEnv()

	if os.Getenv("REMOTE") == "true" {
		b.isRemote = true
	}

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
			err := generateHostsfile(ipText, "server")
			if err != nil {
				text := fmt.Sprintf("IP address should be version 4. Kindly put again")
				b.SendMsg(b.ID, text)
				continue
			}
			text := fmt.Sprintf("Your server IP is %s ", ipText)
			b.SendMsg(b.ID, text)
			b.turnOutMode("saveIP")
			text2 := fmt.Sprintf("Please let me know your ssh private key, (only stored into this machine)")
			b.SendMsg(b.ID, text2)
			continue
		}
		if b.mode == "saveIP" {
			sshKey := update.Message.Text
			err := generateSSHKeyfile(sshKey)
			if err != nil {
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
		if update.Message.Text == "/setup_bot" {
			if b.isRemote {
				continue
			}
			err := b.updateHostAndKeys()
			if err != nil {
				log.Info(err)
				continue
			}
			extVars := map[string]string{
				"BOT_TOKEN": b.bot.Token,
				"CHAT_ID":   strconv.Itoa(int(b.ID)),
				"SSH_KEY":   b.sshKey,
				"IP_ADDR":   b.nodeIP,
				"REMOTE":    "true",
			}
			err = b.execAnsible("./playbooks/bot_install.yml", extVars)
			if err != nil {
				log.Info(err)
				continue
			}
			log.Fatal("Bot is moved out to your server!")
			b.isRemote = true
			continue
		}
		if update.Message.Text == "/setup_infura" {
			extVars := map[string]string{}
			err = b.execAnsible("./playbooks/testnet_infura.yml", extVars)
			if err != nil {
				log.Info(err)
				continue
			}
		}
		if update.Message.Text == "/setup_swingby_node" {
			command := fmt.Sprintf(
				"%s && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -v -i %s %s --extra-vars ''",
				sshAgtCMD,
				hostsFilePath,
				"./playbooks/testnet_tbtc_goerli.yml")
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

func (b *Bot) loadBotEnv() {
	if os.Getenv("CHAT_ID") != "" {
		intID, err := strconv.Atoi(os.Getenv("CHAT_ID"))
		if err == nil {
			b.ID = int64(intID)
		}
		log.Infof("chatID=%d", b.ID)
	}
	if os.Getenv("IP_ADDR") != "" {
		generateHostsfile(os.Getenv("IP_ADDR"), "server")
		log.Infof("IP_ADDR=%s", os.Getenv("IP_ADDR"))
	}
	if os.Getenv("SSH_KEY") != "" {
		generateSSHKeyfile(os.Getenv("SSH_KEY"))
		log.Info("ssh key loaded")
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

func (b *Bot) SendMsg(id int64, text string) {
	msg := tgbotapi.NewMessage(id, text)
	msg.ParseMode = "HTML"
	b.bot.Send(msg)
}

func (b *Bot) updateHostAndKeys() error {
	host, err := getFileHostfile()
	if err != nil {
		return err
	}
	b.nodeIP = host
	log.Infof("loaded IP form file: %s", host)
	// load ssh key file
	key, err := getFileSSHKeyfie()
	if err != nil {
		return err
	}
	b.sshKey = key
	return nil
}

func (b *Bot) execAnsible(playbookPath string, extVars map[string]string) error {
	err := b.updateHostAndKeys()
	ansiblePlaybookConnectionOptions := &ansibler.AnsiblePlaybookConnectionOptions{
		AskPass:    false,
		PrivateKey: sshKeyFilePath,
		Timeout:    "30",
	}

	ansiblePlaybookOptions := &ansibler.AnsiblePlaybookOptions{
		Inventory: "./data/hosts",
	}
	for keyVar, valueVar := range extVars {
		ansiblePlaybookOptions.AddExtraVar(keyVar, valueVar)
	}
	playbook := &ansibler.AnsiblePlaybookCmd{
		Playbook:          playbookPath,
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		//StdoutCallback:    "json",
	}
	log.Info(playbook.String())
	err = playbook.Run()
	if err != nil {
		return err
	}
	return nil
}

func generateHostsfile(nodeIP string, target string) error {
	ipAddr := net.ParseIP(nodeIP)
	if ipAddr == nil {
		return errors.New("IP addr error")
	}
	text := fmt.Sprintf("[%s]\n%s", target, nodeIP)
	err := ioutil.WriteFile(hostsFilePath, []byte(text), 0666)
	if err != nil {
		return err
	}
	return nil
}

func getFileHostfile() (string, error) {
	str, err := ioutil.ReadFile(hostsFilePath)
	if err != nil {
		return "", err
	}
	strs := strings.Split(string(str), "]")
	ipAddr := net.ParseIP(strs[1][1:])
	if ipAddr == nil {
		return "", errors.New("IP addr parse error")
	}
	return ipAddr.String(), nil
}

func getFileSSHKeyfie() (string, error) {
	// if err := verifySSHkey() != nil {
	// 	return "", errors.New("ssh key is invalid")
	// }
	str, err := ioutil.ReadFile(sshKeyFilePath)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func generateSSHKeyfile(key string) error {
	// if verifySSHkey() != nil {
	// 	return errors.New("ssh key is invalid")
	// }
	text := fmt.Sprintf("%s\n", key)
	err := ioutil.WriteFile(sshKeyFilePath, []byte(text), 0600)
	if err != nil {
		return err
	}
	return nil
}

func verifySSHkey() error {
	cmd := exec.Command("sh", "-c", sshAgtCMD)
	log.Info(cmd)
	err := cmd.Run()
	return err
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
