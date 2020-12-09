package bot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	ansibler "github.com/apenella/go-ansible"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	dataPath     = "./data"
	network1     = "testnet_tbtc_bc"
	network2     = "testnet_tbtc_goerli"
	network3     = "testnet_tbtc_bsc"
	network4     = "mainnet_btc_bc"
	network5     = "mainnet_btc_eth"
	network6     = "mainnet_btc_bsc"
	blockBookBTC = "51.15.143.55:9130"
	blockBookETH = "51.15.143.55:9131"
)

var networks = map[string]string{
	"1": network1,
	"2": network2,
	"3": network3,
	"4": network4,
	"5": network5,
	"6": network6,
}

type Bot struct {
	mu               *sync.RWMutex
	bot              *tgbotapi.BotAPI
	Messages         map[int]string
	ID               int64
	hostUser         string
	nodeIP           string
	sshKey           string
	network          string
	moniker          string
	bootstrapNode    string
	coinA            string
	coinB            string
	rewardAddressBTC string
	rewardAddressETH string
	rewardAddressBNB string
	blockBookBTC     string
	blockBookETH     string
	stakeAddr        string
	stakeTx          string
	keygenUntil      string
	isRemote         bool
	isTestnet        bool
}

func NewBot(token string) (*Bot, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	initTime := time.Date(2014, time.December, 31, 12, 13, 24, 0, time.UTC)
	bot := &Bot{
		mu:            new(sync.RWMutex),
		Messages:      make(map[int]string),
		bot:           b,
		ID:            0,
		coinA:         "BTC",
		coinB:         "BTCB",
		blockBookBTC:  blockBookBTC,
		blockBookETH:  blockBookETH,
		keygenUntil:   initTime.Format(time.RFC3339),
		bootstrapNode: "https://tbtc-goerli-1.swingby.network",
		network:       networks["1"],
		moniker:       "Default Node",
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

	b.loadHostAndKeys()

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}
	for update := range updates {
		log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)
		if update.Message.Text == "/start" {
			if b.ID == 0 {
				b.ID = update.Message.Chat.ID
			}
			b.SendMsg(b.ID, makeHelloText(), false)
			continue
		}
		if !b.validateChat(update.Message.Chat.ID) {
			continue
		}
		// Handle for reply messages
		if update.Message.ReplyToMessage != nil {
			msg := update.Message.Text
			prevMsg := update.Message.ReplyToMessage
			mode := b.Messages[prevMsg.MessageID]
			if mode == "setup_node_set_network" {
				b.updateNetwork(msg)
				continue
			}
			if mode == "setup_node_moniker" {
				b.updateNodeMoniker(msg)
			}
			if mode == "setup_node_btc_addr" {
				b.updateBTCAddr(msg)
				continue
			}
			if mode == "setup_node_bnb_addr" {
				b.updateBNBAddr(msg)
				continue
			}
			if mode == "setup_node_eth_addr" {
				b.updateETHAddr(msg)
				continue
			}
			if mode == "setup_node_stake_tx" {
				b.updateStakeTx(msg)
				continue
			}

			// Set node config
			if mode == "setup_config_1" {
				err := generateHostsfile(msg, "server")
				if err != nil {
					text := fmt.Sprintf("IP address should be version 4. Kindly put again")
					newMsg, _ := b.SendMsg(b.ID, text, true)
					b.Messages[newMsg.MessageID] = "setup_config_1"
					continue
				}
				err = b.loadHostAndKeys()
				if err != nil {
					text := fmt.Sprintf("Config are something wrong. please try again")
					b.SendMsg(b.ID, text, false)
					continue
				}
				b.SendMsg(b.ID, doneSSHKeyText(), false)
				//b.Messages[newMsg.MessageID] = "setup_config_2"
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

		if update.Message.Text == "/setup_config" {
			// Disable if remote is `true`
			if b.isRemote {
				continue
			}
			msg, err := b.SendMsg(b.ID, makeHostText(), true)
			if err != nil {
				continue
			}
			b.Messages[msg.MessageID] = "setup_config_1"
			continue
		}

		if update.Message.Text == "/setup_your_bot" {
			// Disable if remote is `true`
			if b.isRemote {
				continue
			}
			err := b.loadHostAndKeys()
			if err != nil {
				log.Info(err)
				continue
			}
			b.SendMsg(b.ID, makeDeployBotMessage(), false)
			extVars := map[string]string{
				"USER":      b.hostUser,
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
			log.Info("Bot is moved out to your server!")
			os.Exit(0)
			continue
		}
		if update.Message.Text == "/deploy_infura" {
			extVars := map[string]string{}
			b.SendMsg(b.ID, makeDeployInfuraMessage(), false)
			targetPath := "./playbooks/testnet_infura.yml"
			if b.network == networks["3"] || b.network == networks["4"] || b.network == networks["5"] {
				targetPath = "./playbooks/mainet_infura.yml"
			}
			err = b.execAnsible(targetPath, extVars)
			if err != nil {
				log.Info(err)
				continue
			}
			b.SendMsg(b.ID, doneDeployInfuraMessage(), false)
			continue
		}

		if update.Message.Text == "/setup_node" {
			msg, err := b.SendMsg(b.ID, b.makeNodeText(), true)
			if err != nil {
				continue
			}
			b.Messages[msg.MessageID] = "setup_node_set_network"
			continue
		}
		if update.Message.Text == "/deploy_node" {
			extVars := map[string]string{
				"TAG":            "latest",
				"IP_ADDR":        b.nodeIP,
				"BOOTSTRAP_NODE": b.bootstrapNode,
				"K_UNTIL":        b.keygenUntil,
			}
			path := fmt.Sprintf("./playbooks/%s.yml", b.network)
			err = b.execAnsible(path, extVars)
			if err != nil {
				log.Error(err)
				b.SendMsg(b.ID, errorDeployBotMessage(), false)
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

func (b *Bot) updateStakeTx(msg string) {
	stakeTx := msg
	if stakeTx == "" {
		text := fmt.Sprintf("stakeTx not exist, Please type again")
		newMsg, _ := b.SendMsg(b.ID, text, true)
		b.Messages[newMsg.MessageID] = "setup_node_stake_tx"
		return
	}
	b.stakeTx = stakeTx
	path := fmt.Sprintf("%s/%s", dataPath, b.network)
	b.storeConfig(path, 15, 25)
	b.SendMsg(b.ID, doneConfigGenerateText(), false)
}

func (b *Bot) updateETHAddr(msg string) {
	address := msg
	check := b.checkInput(address, "setup_node_eth_addr")
	if check == 0 {
		return
	}
	if check == 1 {
		b.rewardAddressETH = address
	}
	b.SendMsg(b.ID, b.makeStoreKeyText(), false)
	rewardAddr := ""
	if b.network == network1 {
		rewardAddr = b.rewardAddressBTC
		b.isTestnet = true
		b.coinB = "BTCB"
	}
	if b.network == network2 {
		rewardAddr = b.rewardAddressETH
		b.isTestnet = true
		b.coinB = "BTCE"
	}
	if b.network == network3 {
		rewardAddr = b.rewardAddressETH
		b.isTestnet = true
		b.coinB = "BTCK"
	}
	path := fmt.Sprintf("%s/%s", dataPath, b.network)
	memo, err := b.generateKeys(path, rewardAddr, b.isTestnet)
	if err != nil {
		log.Info(err)
		return
	}
	b.sendKeyStoreFile(path)
	b.SendMsg(b.ID, makeStakeTxText(b.stakeAddr, memo), false)
	newMsg, _ := b.SendMsg(b.ID, askStakeTxText(), true)
	b.Messages[newMsg.MessageID] = "setup_node_stake_tx"
}

func (b *Bot) updateBNBAddr(msg string) {
	address := msg
	check := b.checkInput(address, "setup_node_bnb_addr")
	if check == 0 {
		return
	}
	if check == 1 {
		b.rewardAddressBNB = address
	}
	newMsg, _ := b.SendMsg(b.ID, b.makeRewardAddressETH(), true)
	b.Messages[newMsg.MessageID] = "setup_node_eth_addr"
}

func (b *Bot) updateBTCAddr(msg string) {
	address := msg
	check := b.checkInput(address, "setup_node_btc_addr")
	if check == 0 {
		return
	}
	if check == 1 {
		b.rewardAddressBTC = address
	}
	newMsg, _ := b.SendMsg(b.ID, b.makeRewardAddressBNB(), true)
	b.Messages[newMsg.MessageID] = "setup_node_bnb_addr"
}

func (b *Bot) updateNodeMoniker(msg string) {
	moniker := msg
	check := b.checkInput(moniker, "setup_node_moniker")
	if check == 0 {
		return
	}
	if check == 1 {
		b.moniker = moniker
	}
	newMsg, _ := b.SendMsg(b.ID, b.makeRewardAddressBTC(), true)
	b.Messages[newMsg.MessageID] = "setup_node_btc_addr"
}

func (b *Bot) updateNetwork(msg string) {
	network := networks[msg]
	check := b.checkInput(network, "setup_node_set_network")
	if check == 0 {
		return
	}
	if check == 1 {
		b.network = network
	}
	newMsg, _ := b.SendMsg(b.ID, b.makeUpdateMoniker(), true)
	b.Messages[newMsg.MessageID] = "setup_node_moniker"
}

func (b *Bot) checkInput(input string, keepMode string) int {
	if input == "" {
		text := fmt.Sprintf("input is wrong, Please type again")
		newMsg, _ := b.SendMsg(b.ID, text, true)
		b.Messages[newMsg.MessageID] = keepMode
		return 0
	}
	if input == "none" {
		return 2
	}
	return 1
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
		log.Infof("IP address stored IP_ADDR=%s", os.Getenv("IP_ADDR"))
	}
	if os.Getenv("SSH_KEY") != "" {
		generateSSHKeyfile(os.Getenv("SSH_KEY"))
		log.Info("A ssh key stored")
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

func (b *Bot) loadHostAndKeys() error {
	host, err := getFileHostfile()
	if err != nil {
		return err
	}
	b.nodeIP = host
	b.blockBookBTC = fmt.Sprintf("%s:9130", b.nodeIP)
	b.blockBookETH = fmt.Sprintf("%s:9131", b.nodeIP)

	log.Infof("Loaded IP form file: %s", host)
	// load ssh key file
	key, err := getFileSSHKeyfie()
	if err != nil {
		return err
	}
	log.Infof("Loaded ssh keys")
	b.sshKey = key
	return nil
}

func (b *Bot) sendKeyStoreFile(path string) {
	stakeKeyPath := fmt.Sprintf("%s/key_%s.json", path, b.network)
	msg := tgbotapi.NewDocumentUpload(b.ID, stakeKeyPath)
	b.bot.Send(msg)
}

func (b *Bot) execAnsible(playbookPath string, extVars map[string]string) error {
	sshKeyFilePath := fmt.Sprintf("%s/ssh_key", dataPath)
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
	}
	log.Info(playbook.String())
	err := playbook.Run()
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
	path := fmt.Sprintf("%s/hosts", dataPath)
	err := ioutil.WriteFile(path, []byte(text), 0666)
	if err != nil {
		return err
	}
	return nil
}

func getFileHostfile() (string, error) {
	path := fmt.Sprintf("%s/hosts", dataPath)
	str, err := ioutil.ReadFile(path)
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
	path := fmt.Sprintf("%s/ssh_key", dataPath)
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func generateSSHKeyfile(key string) error {
	text := fmt.Sprintf("%s\n", key)
	path := fmt.Sprintf("%s/ssh_key", dataPath)
	err := ioutil.WriteFile(path, []byte(text), 0600)
	if err != nil {
		return err
	}
	return nil
}
