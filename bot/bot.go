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
	network1       = "testnet_tbtc_bc"
	network2       = "testnet_tbtc_goerli"
	network3       = "testnet_tbtc_bsc"
)

type Bot struct {
	mu               *sync.RWMutex
	bot              *tgbotapi.BotAPI
	Messages         map[int]string
	ID               int64
	nodeIP           string
	sshKey           string
	network          string
	coinA            string
	coinB            string
	rewardAddressBTC string
	rewardAddressETH string
	rewardAddressBNB string
	blockBookBTC     string
	blockBookETH     string
	stakeTx          string
	isRemote         bool
}

func NewBot(token string) (*Bot, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot := &Bot{
		mu:       new(sync.RWMutex),
		Messages: make(map[int]string),
		bot:      b,
		ID:       0,
		coinA:    "BTC",
		coinB:    "BTCE",
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

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}
	for update := range updates {
		log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)
		if !b.validateChat(update.Message.Chat.ID) && b.ID != 0 {
			continue
		}
		// Handle for reply messages
		if update.Message.ReplyToMessage != nil {
			msg := update.Message.Text
			prevMsg := update.Message.ReplyToMessage
			mode := b.Messages[prevMsg.MessageID]
			if mode == "setup_node_set_network" {
				network := networks[msg]
				if network == "" {
					text := fmt.Sprintf("network is not exist, Please type again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_node_set_network"
					continue
				}
				b.network = network
				newMsg, _ := b.SendMsg(b.ID, makeRewardAddressBTC(), true)
				b.Messages[newMsg.MessageID] = "setup_node_btc_addr"
				continue
			}
			if mode == "setup_node_btc_addr" {
				address := msg
				if address == "" {
					text := fmt.Sprintf("address not exist, Please type again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_node_btc_addr"
					continue
				}
				b.rewardAddressBTC = address
				newMsg, _ := b.SendMsg(b.ID, makeRewardAddressBNB(), true)
				b.Messages[newMsg.MessageID] = "setup_node_bnb_addr"
				continue
			}
			if mode == "setup_node_bnb_addr" {
				address := msg
				if address == "" {
					text := fmt.Sprintf("address not exist, Please type again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_node_bnb_addr"
					continue
				}
				b.rewardAddressBNB = address
				newMsg, _ := b.SendMsg(b.ID, makeRewardAddressETH(), true)
				b.Messages[newMsg.MessageID] = "setup_node_eth_addr"
				continue
			}
			if mode == "setup_node_eth_addr" {
				address := msg
				if address == "" {
					text := fmt.Sprintf("address not exist, Please type again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_node_eth_addr"
					continue
				}
				b.rewardAddressETH = address
				b.SendMsg(b.ID, makeStoreKeyText(), false)
				rewardAddr := ""
				isTestnet := true
				if b.network == network1 {
					rewardAddr = b.rewardAddressBTC
				}
				if b.network == network2 {
					rewardAddr = b.rewardAddressETH
				}
				if b.network == network3 {
					rewardAddr = b.rewardAddressETH // BSC
				}
				addr, memo := generateKeys("./data", rewardAddr, isTestnet)
				b.SendMsg(b.ID, makeStakeTxText(addr, memo), false)
				newMsg, _ := b.SendMsg(b.ID, askStakeTxText(), true)
				b.Messages[newMsg.MessageID] = "setup_node_stake_tx"
				continue
			}
			if mode == "setup_node_stake_tx" {
				stakeTx := msg
				if stakeTx == "" {
					text := fmt.Sprintf("stakeTx not exist, Please type again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_node_stake_tx"
					continue
				}
				b.stakeTx = stakeTx
				b.storeConfig("./data", "testMoniker")
				b.SendMsg(b.ID, doneConfigGenerateText(), false)
				continue
			}
			if mode == "setup_config_1" {
				err := generateHostsfile(msg, "server")
				if err != nil {
					text := fmt.Sprintf("IP address should be version 4. Kindly put again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_config_1"
					continue
				}
				newMsg, _ := b.SendMsg(b.ID, seutpSSHKeyText(msg), true)
				b.Messages[newMsg.MessageID] = "setup_config_2"
				continue
			}
			if mode == "setup_config_2" {
				err := generateSSHKeyfile(msg)
				if err != nil {
					text := fmt.Sprintf("SSH priv key is not valid. Kindly put again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_config_2"
					continue
				}
				b.SendMsg(b.ID, doneSSHKeyText(), true)
				continue
			}
		}

		if update.Message.Text == "/start" {
			if b.ID == 0 {
				b.ID = update.Message.Chat.ID
			}
			b.SendMsg(b.ID, makeHelloText(), false)
			continue
		}
		if update.Message.Text == "/setup_config" {
			msg, err := b.SendMsg(b.ID, makeHostText(), true)
			if err != nil {
				continue
			}
			b.Messages[msg.MessageID] = "setup_config_1"
			continue
		}

		if update.Message.Text == "/setup_node" {
			msg, err := b.SendMsg(b.ID, makeNodeText(), true)
			log.Info(err)
			if err != nil {
				continue
			}
			b.Messages[msg.MessageID] = "setup_node_set_network"
			continue
		}

		if update.Message.Text == "/setup_your_bot" {
			if b.isRemote {
				continue
			}
			err := b.updateHostAndKeys()
			if err != nil {
				log.Info(err)
				continue
			}
			b.SendMsg(b.ID, makeDeployBotMessage(), false)
			extVars := map[string]string{
				"BOT_TOKEN": b.bot.Token,
				"CHAT_ID":   strconv.Itoa(int(b.ID)),
				"SSH_KEY":   b.sshKey,
				"IP_ADDR":   b.nodeIP,
				"REMOTE":    "true",
			}
			err = b.execAnsible("./playbooks/bot_install.yml", extVars)
			if err != nil {
				log.Error(err)
				b.SendMsg(b.ID, errorDeployBotMessage(), false)
				continue
			}
			b.SendMsg(b.ID, doneDeployBotMessage(), false)
			log.Panicf("Bot is moved out to your server!")
			continue
		}
		if update.Message.Text == "/setup_infura" {
			extVars := map[string]string{}
			b.SendMsg(b.ID, makeDeployInfuraMessage(), false)
			err = b.execAnsible("./playbooks/testnet_infura.yml", extVars)
			if err != nil {
				log.Info(err)
				continue
			}
			b.SendMsg(b.ID, doneDeployInfuraMessage(), false)
			continue
		}
		if update.Message.Text == "/deploy_node" {
			extVars := map[string]string{}
			b.SendMsg(b.ID, makeDeployNodeMessage(), false)
			err = b.execAnsible("./playbooks/testnet_node.yml", extVars)
			if err != nil {
				log.Info(err)
				continue
			}
			b.SendMsg(b.ID, doneDeployNodeMessage(), false)
		}
		// Default response of say hi
		if update.Message.Text == "hi" || update.Message.Text == "Hi" {
			b.SendMsg(b.ID, `Start with /start`, false)
		}
	}
}

func (b *Bot) storeConfig(path string, moniker string) {
	storeConfig(path, moniker, 15, 25, b.coinA, b.coinB, b.blockBookBTC, b.blockBookETH, b.rewardAddressBTC, b.rewardAddressETH, b.rewardAddressBNB, b.stakeTx)
}

func (b *Bot) loadBotEnv() {
	if os.Getenv("REMOTE") == "true" {
		b.isRemote = true
	}
	if os.Getenv("CHAT_ID") != "" {
		intID, err := strconv.Atoi(os.Getenv("CHAT_ID"))
		if err == nil {
			b.ID = int64(intID)
		}
		log.Infof("ChatID=%d", b.ID)
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

func (b *Bot) SendMsg(id int64, text string, isReply bool) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(id, text)
	if isReply {
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
	}
	msg.ParseMode = "HTML"
	return b.bot.Send(msg)
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

type Executor struct {
}

func (e *Executor) Execute(command string, args []string, prefix string) error {
	cmd := exec.Command(command, args...)
	err := cmd.Run()
	if err != nil {
		return errors.New("(DefaultExecute::Execute) -> " + err.Error())
	}
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
		Exec:              &Executor{},
		StdoutCallback:    "json",
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
