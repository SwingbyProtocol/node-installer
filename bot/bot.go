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

	"github.com/SwingbyProtocol/tx-indexer/api"
	ansibler "github.com/apenella/go-ansible"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	version     = "v0.1.0"
	dataPath    = "./data"
	network1    = "mainnet_btc_eth"
	network2    = "mainnet_btc_bc"
	network3    = "testnet_tbtc_goerli"
	network4    = "testnet_tbtc_bc"
	maxDataSize = 817083983700
)

var Networks = map[string]string{
	"1": network1,
	"2": network2,
	"3": network3,
	"4": network4,
}

type Bot struct {
	Messages           map[int]string
	ID                 int64
	mu                 *sync.RWMutex
	bot                *tgbotapi.BotAPI
	api                *api.Resolver
	version            string
	hostUser           string
	nodeIP             string
	containerName      string
	sshKey             string
	nConf              *NodeConfig
	isRemote           bool
	isLocked           bool
	isConfirmed        map[string]bool
	bestHeightBTC      int
	stuckCountBTC      int
	bestHeightETH      int
	stuckCountETH      int
	isSyncedBTC        bool
	isSyncedETH        bool
	isSyncedMempoolBTC bool
	isSyncedMempoolETH bool
	syncProgress       float64
	syncBTCRatio       float64
	syncETHRatio       float64
}

func NewBot(token string) (*Bot, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot := &Bot{
		Messages:      make(map[int]string),
		ID:            0,
		mu:            new(sync.RWMutex),
		bot:           b,
		api:           api.NewResolver("", 200),
		version:       version,
		hostUser:      "root",
		containerName: "node_installer",
		nConf:         NewNodeConfig(),
		isConfirmed:   make(map[string]bool),
	}
	return bot, nil
}

func (b *Bot) Start() {
	b.bot.Debug = false
	b.api.SetTimeout(15 * time.Second)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	log.Infof("Authorized on account %s", b.bot.Self.UserName)

	b.loadSystemEnv()
	b.loadHostAndKeys()
	b.nConf.loadConfig()

	log.Infof("Now keygenUntil is %s", b.nConf.KeygenUntil)

	ticker := time.NewTicker(30 * time.Second)
	b.checkBlockBooks()
	go func() {
		for {
			<-ticker.C
			b.checkBlockBooks()
		}
	}()

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		log.Error(err)
		return
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.From == nil {
			continue
		}
		log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)

		commands := strings.Split(update.Message.Text, "@")
		cmd := commands[0]

		if cmd == "/start" {
			if b.ID == 0 {
				b.ID = update.Message.Chat.ID
				b.SendMsg(b.ID, b.makeHelloText(), false, false)
				continue
			}
			if !b.validateChat(update.Message.Chat.ID) {
				continue
			}
			b.SendMsg(b.ID, b.makeHelloText(), false, false)
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
			// if mode == "setup_node_btc_addr" {
			// 	b.updateBTCAddr(msg)
			// 	continue
			// }
			// if mode == "setup_node_bnb_addr" {
			// 	b.updateBNBAddr(msg)
			// 	continue
			// }
			if mode == "setup_node_eth_addr" {
				b.updateETHAddr(msg)
				continue
			}
			if mode == "setup_node_stake_addr" {
				b.updateStakeAddr(msg)
				continue
			}
			// Set server configs
			if mode == "setup_ip_addr" {
				b.setupIPAddr(msg)
				continue
			}
			if mode == "setup_domain" {
				b.setupDomain(msg)
				continue
			}
			if mode == "setup_username" {
				b.setupUser(msg)
				continue
			}
		}

		if cmd == "/setup_server_config" {
			// Disable if remote is `true`
			if b.isRemote {
				continue
			}
			msg, err := b.SendMsg(b.ID, b.makeSetupIPText(), true, false)
			if err != nil {
				continue
			}
			b.Messages[msg.MessageID] = "setup_ip_addr"
			continue
		}

		if cmd == "/setup_domain" {
			newMsg, err := b.SendMsg(b.ID, b.setupDomainText(), true, false)
			if err != nil {
				continue
			}
			b.Messages[newMsg.MessageID] = "setup_domain"
			continue
		}

		if cmd == "/setup_your_bot" {
			// Disable if remote is `true`
			if b.isRemote {
				continue
			}
			err := b.loadHostAndKeys()
			if err != nil {
				log.Info(err)
				continue
			}
			if b.checkProcess() {
				continue
			}
			b.SendMsg(b.ID, makeDeployBotMessage(), false, false)
			extVars := map[string]string{
				"CONT_NAME": b.containerName,
				"HOST_USER": b.hostUser,
				"BOT_TOKEN": b.bot.Token,
				"CHAT_ID":   strconv.Itoa(int(b.ID)),
				"SSH_KEY":   b.sshKey,
				"IP_ADDR":   b.nodeIP,
				"REMOTE":    "true",
			}
			onSuccess := func() {
				b.SendMsg(b.ID, doneDeployBotMessage(), false, false)
				log.Info("Bot is moved out to your server!")
				b.cooldown()
				os.Exit(0)
			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorDeployBotMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible("./playbooks/bot_install.yml", extVars, onSuccess, onError)
			continue
		}

		if cmd == "/setup_node" {
			msg, err := b.SendMsg(b.ID, b.makeNodeText(), true, false)
			if err != nil {
				continue
			}
			b.Messages[msg.MessageID] = "setup_node_set_network"
			continue
		}

		if cmd == "/upgrade_your_bot" {
			if !b.isRemote {
				continue
			}
			err := b.loadHostAndKeys()
			if err != nil {
				log.Info(err)
				continue
			}
			if b.checkProcess() {
				continue
			}
			b.SendMsg(b.ID, makeUpgradeBotMessage(), false, false)
			contName := b.containerName
			if b.containerName == "node_installer" {
				contName = "node_installer_fork"
			} else {
				contName = "node_installer"
			}
			extVars := map[string]string{
				"CONT_NAME": contName,
				"HOST_USER": b.hostUser,
				"BOT_TOKEN": b.bot.Token,
				"CHAT_ID":   strconv.Itoa(int(b.ID)),
				"SSH_KEY":   b.sshKey,
				"IP_ADDR":   b.nodeIP,
				"REMOTE":    "true",
			}
			onSuccess := func() {
				b.SendMsg(b.ID, b.doneUpgradeBotMessage(), false, false)
				b.cooldown()
				os.Exit(0)
				// extVars := map[string]string{
				// 	"CONT_NAME": b.containerName,
				// 	"HOST_USER": b.hostUser,
				// }
				// b.execAnsible("./playbooks/bot_remove.yml", extVars, nil, nil)
			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorDeployBotMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible("./playbooks/bot_install.yml", extVars, onSuccess, onError)
			continue
		}

		if cmd == "/setup_infura" {
			if b.checkProcess() {
				continue
			}
			if !b.isConfirmed["setup_infura"] {
				b.SendMsg(b.ID, confirmSetupInfuraMessage(), false, false)
				b.mu.Lock()
				b.isConfirmed["setup_infura"] = true
				b.mu.Unlock()
				go func() {
					time.Sleep(10 * time.Second)
					b.mu.Lock()
					b.isConfirmed["setup_infura"] = false
					b.mu.Unlock()
				}()
				b.cooldown()
				continue
			}
			b.mu.Lock()
			b.isConfirmed["setup_infura"] = false
			b.mu.Unlock()
			extVars := map[string]string{
				"HOST_USER": b.hostUser,
			}
			b.SendMsg(b.ID, makeSetupInfuraMessage(), false, false)
			targetPath := "./playbooks/mainnet_infura_setup.yml"
			if b.nConf.IsTestnet {
				targetPath = "./playbooks/testnet_infura_setup.yml"
			}
			onSuccess := func() {
				b.SendMsg(b.ID, doneSetupInfuraMessage(), false, false)
				b.cooldown()
			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorSetupInfuraMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible(targetPath, extVars, onSuccess, onError)
			continue
		}

		if cmd == "/deploy_infura" {
			if b.checkProcess() {
				continue
			}
			if b.syncProgress < 99.99 {
				b.SendMsg(b.ID, rejectDeployInfuraMessage(), false, false)
				b.cooldown()
				continue
			}
			if !b.isConfirmed["deploy_infura"] {
				b.SendMsg(b.ID, confirmDeployInfuraMessage(), false, false)
				b.mu.Lock()
				b.isConfirmed["deploy_infura"] = true
				b.mu.Unlock()
				go func() {
					time.Sleep(10 * time.Second)
					b.mu.Lock()
					b.isConfirmed["deploy_infura"] = false
					b.mu.Unlock()
				}()
				b.cooldown()
				continue
			}
			b.mu.Lock()
			b.isConfirmed["deploy_infura"] = false
			b.mu.Unlock()
			extVars := map[string]string{
				"HOST_USER": b.hostUser,
			}
			b.SendMsg(b.ID, makeDeployInfuraMessage(), false, false)
			targetPath := "./playbooks/mainnet_infura.yml"
			if b.nConf.IsTestnet {
				targetPath = "./playbooks/testnet_infura.yml"
			}
			onSuccess := func() {
				b.SendMsg(b.ID, doneDeployInfuraMessage(), false, false)
				b.cooldown()
			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorDeployInfuraMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible(targetPath, extVars, onSuccess, onError)
			continue
		}

		if cmd == "/check_status" {
			if b.checkProcess() {
				continue
			}
			extVars := map[string]string{
				"HOST_USER": b.hostUser,
				"IP_ADDR":   b.nodeIP,
			}
			b.SendMsg(b.ID, makeCheckNodeMessage(), false, false)
			path := fmt.Sprintf("./playbooks/mainnet_check.yml")
			onSuccess := func() {
				syncDataSize, _ := getDirSizeFromFile()
				parcent := 100 * float64(syncDataSize) / float64(maxDataSize)
				if parcent >= 100 {
					b.syncProgress = 99.99
				}
				if b.isSyncedBTC && b.isSyncedETH {
					b.syncProgress = 100.00
				}
				b.SendMsg(b.ID, b.checkNodeMessage(), false, false)
				b.cooldown()
			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorCheckNodeMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible(path, extVars, onSuccess, onError)
			continue
		}

		if cmd == "/deploy_node" {
			if b.checkProcess() {
				continue
			}
			if b.syncProgress <= 99.99 {
				b.SendMsg(b.ID, rejectDeployNodeMessage(), false, false)
				b.cooldown()
				continue
			}
			extVars := map[string]string{
				"HOST_USER":        b.hostUser,
				"TAG":              "test1",
				"BOOTSTRAP_NODE_1": b.nConf.BootstrapNode[0],
				"BOOTSTRAP_NODE_2": b.nConf.BootstrapNode[1],
				"BOOTSTRAP_NODE_3": b.nConf.BootstrapNode[2],
				"K_UNTIL":          b.nConf.KeygenUntil,
				"LOG_LEVEL":        "INFO",
			}
			b.SendMsg(b.ID, makeDeployNodeMessage(), false, false)
			path := fmt.Sprintf("./playbooks/%s.yml", b.nConf.Network)
			onSuccess := func() {
				b.SendMsg(b.ID, doneDeployNodeMessage(), false, false)
				b.cooldown()
			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorDeployNodeMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible(path, extVars, onSuccess, onError)
			continue
		}

		if cmd == "/deploy_node_debug" {
			if b.checkProcess() {
				continue
			}
			if b.syncProgress <= 99.99 {
				b.SendMsg(b.ID, rejectDeployNodeMessage(), false, false)
				b.cooldown()
				continue
			}
			extVars := map[string]string{
				"HOST_USER":        b.hostUser,
				"TAG":              "test1",
				"BOOTSTRAP_NODE_1": b.nConf.BootstrapNode[0],
				"BOOTSTRAP_NODE_2": b.nConf.BootstrapNode[1],
				"BOOTSTRAP_NODE_3": b.nConf.BootstrapNode[2],
				"K_UNTIL":          b.nConf.KeygenUntil,
				"LOG_LEVEL":        "DEBUG",
			}
			b.SendMsg(b.ID, makeDeployNodeMessage(), false, false)
			path := fmt.Sprintf("./playbooks/%s.yml", b.nConf.Network)
			onSuccess := func() {
				b.SendMsg(b.ID, doneDeployNodeMessage(), false, false)
				b.cooldown()
			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorDeployNodeMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible(path, extVars, onSuccess, onError)
			continue
		}

		if cmd == "/enable_domain" {
			if b.checkProcess() {
				continue
			}
			extVars := map[string]string{
				"HOST_USER": b.hostUser,
				"DOMAIN":    b.nConf.Domain,
			}
			b.SendMsg(b.ID, b.makeDomainMessage(), false, false)
			path := fmt.Sprintf("./playbooks/enable_domain.yml")
			onSuccess := func() {
				b.SendMsg(b.ID, b.doneDomainMessage(), false, false)
				b.cooldown()

			}
			onError := func(err error) {
				log.Error(err)
				b.SendMsg(b.ID, errorDomainMessage(), false, false)
				b.cooldown()

			}
			b.execAnsible(path, extVars, onSuccess, onError)
			continue
		}
		// Default response of say hi
		if cmd == "hi" || cmd == "Hi" {
			b.SendMsg(b.ID, `Let's start with /start`, false, false)
		}
	}
}

func (b *Bot) checkProcess() bool {
	b.mu.RLock()
	if b.isLocked {
		b.mu.RUnlock()
		text := fmt.Sprintf("Process is already started")
		b.SendMsg(b.ID, text, false, false)
		return true
	}
	b.mu.RUnlock()
	b.mu.Lock()
	b.isLocked = true
	b.mu.Unlock()
	return false
}

func (b *Bot) cooldown() {
	b.mu.Lock()
	b.isLocked = false
	b.mu.Unlock()
}
func (b *Bot) setupIPAddr(msg string) {
	err := generateHostsfile(msg, "server")
	if err != nil {
		text := fmt.Sprintf("IP address should be version 4. Kindly put again")
		newMsg, _ := b.SendMsg(b.ID, text, true, false)
		b.Messages[newMsg.MessageID] = "setup_ip_addr"
		return
	}
	b.nodeIP = msg
	newMsg, _ := b.SendMsg(b.ID, b.setupIPAndAskUserNameText(), true, false)
	b.Messages[newMsg.MessageID] = "setup_username"
}

func (b *Bot) setupUser(msg string) {
	check := b.checkInput(msg, "setup_username")
	if check == 0 {
		return
	}
	if check == 1 {
		b.hostUser = msg
	}
	err := b.loadHostAndKeys()
	if err != nil {
		text := fmt.Sprintf("SSH_KEY load error. please check data/ssh_key file again")
		b.SendMsg(b.ID, text, false, false)
		return
	}
	b.SendMsg(b.ID, b.setupUsernameAndLoadSSHkeyText(), false, false)
}

func (b *Bot) setupDomain(msg string) {
	check := b.checkInput(msg, "setup_domain")
	if check == 0 {
		return
	}
	if check == 1 {
		b.nConf.SetDomain(msg)
		b.nConf.storeConfig()
		b.nConf.saveConfig()
		b.nConf.loadConfig()
	}
	b.SendMsg(b.ID, b.doneDomainText(), false, false)
}

func (b *Bot) updateStakeAddr(msg string) {
	stakeTx := msg
	check := b.checkInput(stakeTx, "setup_node_stake_addr")
	if check == 0 {
		return
	}
	if check == 1 {
		b.nConf.StakeAddr = msg
	}
	b.nConf.storeConfig()
	b.nConf.saveConfig()
	b.nConf.loadConfig()
	b.SendMsg(b.ID, doneConfigGenerateText(), false, false)
}

func (b *Bot) updateETHAddr(msg string) {
	address := msg
	check := b.checkInput(address, "setup_node_eth_addr")
	if check == 0 {
		return
	}
	if check == 1 {
		b.nConf.RewardAddressETH = address
	}
	b.SendMsg(b.ID, b.makeStoreKeyText(), false, false)
	switch b.nConf.Network {
	case network1:
		b.nConf.SetMainnet()
		b.nConf.SetTSSGroup(10, 31)
		b.nConf.CoinA = "WBTC"
		b.nConf.CoinB = "BTC"
	case network2:
		b.nConf.SetMainnet()
		b.nConf.CoinA = "BTCB"
		b.nConf.CoinB = "BTC"
	case network3:
		b.nConf.SetTestnet()
		b.nConf.SetTSSGroup(50, 25)
		b.nConf.CoinA = "BTCE"
		b.nConf.CoinB = "BTC"
	case network4:
		b.nConf.SetTestnet()
		b.nConf.CoinB = "BTCB"
	}
	path := fmt.Sprintf("%s/%s", dataPath, b.nConf.Network)
	isLoad, err := b.generateKeys(path)
	if err != nil {
		log.Info(err)
		return
	}
	if !isLoad {
		b.sendKeyStoreFile(path)
	}
	b.SendMsg(b.ID, b.makeStakeTxText(), false, true)
	newMsg, _ := b.SendMsg(b.ID, b.askStakeTxText(), true, false)
	b.Messages[newMsg.MessageID] = "setup_node_stake_addr"
}

// func (b *Bot) updateBNBAddr(msg string) {
// 	address := msg
// 	check := b.checkInput(address, "setup_node_bnb_addr")
// 	if check == 0 {
// 		return
// 	}
// 	if check == 1 {
// 		b.nConf.RewardAddressBNB = address
// 	}
// 	newMsg, _ := b.SendMsg(b.ID, b.makeRewardAddressETH(), true)
// 	b.Messages[newMsg.MessageID] = "setup_node_eth_addr"
// }

// func (b *Bot) updateBTCAddr(msg string) {
// 	address := msg
// 	check := b.checkInput(address, "setup_node_btc_addr")
// 	if check == 0 {
// 		return
// 	}
// 	if check == 1 {
// 		b.nConf.RewardAddressBTC = address
// 	}
// 	newMsg, _ := b.SendMsg(b.ID, b.makeRewardAddressBNB(), true)
// 	b.Messages[newMsg.MessageID] = "setup_node_bnb_addr"
// }

func (b *Bot) updateNodeMoniker(msg string) {
	moniker := msg
	check := b.checkInput(moniker, "setup_node_moniker")
	if check == 0 {
		return
	}
	if check == 1 {
		b.nConf.Moniker = moniker
	}
	newMsg, _ := b.SendMsg(b.ID, b.makeRewardAddressETH(), true, false)
	b.Messages[newMsg.MessageID] = "setup_node_eth_addr"
}

func (b *Bot) updateNetwork(msg string) {
	network := Networks[msg]
	check := b.checkInput(network, "setup_node_set_network")
	if check == 0 {
		return
	}
	if check == 1 {
		b.nConf.Network = network
	}
	newMsg, _ := b.SendMsg(b.ID, b.makeUpdateMoniker(), true, false)
	b.Messages[newMsg.MessageID] = "setup_node_moniker"
}

func (b *Bot) checkInput(input string, keepMode string) int {
	if input == "" {
		text := fmt.Sprintf("input is wrong, Please type again")
		newMsg, _ := b.SendMsg(b.ID, text, true, false)
		b.Messages[newMsg.MessageID] = keepMode
		return 0
	}
	if input == "none" {
		return 2
	}
	return 1
}

func (b *Bot) loadSystemEnv() {
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
		err := generateHostsfile(os.Getenv("IP_ADDR"), "server")
		if err == nil {
			log.Infof("IP address stored IP_ADDR=%s", os.Getenv("IP_ADDR"))
		}
	}
	if os.Getenv("CONT_NAME") != "" {
		b.containerName = os.Getenv("CONT_NAME")
		log.Infof("CONT_NAME=%s", b.containerName)
	}
	if os.Getenv("HOST_USER") != "" {
		b.hostUser = os.Getenv("HOST_USER")
		log.Infof("HOST_USER=%s", b.hostUser)
	}
	if os.Getenv("SSH_KEY") != "" {
		generateSSHKeyfile(os.Getenv("SSH_KEY"))
		log.Info("SSH priv key is stored")
	}
}

func (b *Bot) SendMsg(id int64, text string, isReply bool, hidePreview bool) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(id, text)
	if isReply {
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
	}
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = hidePreview
	return b.bot.Send(msg)
}

func (b *Bot) validateChat(chatID int64) bool {
	if b.ID == chatID {
		return true
	}
	return false
}

func (b *Bot) loadHostAndKeys() error {
	host, err := getFileHostfile()
	if err != nil {
		return err
	}
	b.nodeIP = host
	log.Infof("Loaded target IPv4:%s for your server from hosts file.", host)
	// load ssh key file
	key, err := getFileSSHKeyfie()
	if err != nil {
		return err
	}
	log.Infof("Loaded SSH priv key")
	b.sshKey = key
	return nil
}

func (b *Bot) sendKeyStoreFile(path string) {
	stakeKeyPath := fmt.Sprintf("%s/key_%s.json", path, b.nConf.Network)
	msg := tgbotapi.NewDocumentUpload(b.ID, stakeKeyPath)
	b.bot.Send(msg)
}

func (b *Bot) execAnsible(playbookPath string, extVars map[string]string, onSuccess func(), onError func(err error)) {
	ansibler.AnsibleAvoidHostKeyChecking()

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
		Exec:              &BotExecute{},
	}
	go func() {
		log.Info(playbook.String())
		err := playbook.Run()
		if err != nil {
			onError(err)
			return
		}
		onSuccess()
	}()
}

func (b *Bot) checkBlockBooks() {
	resBTC := BlockBook{}
	resETH := BlockBook{}
	uriBTC := fmt.Sprintf("http://%s/api/", b.nConf.BlockBookBTC)
	b.api.GetRequest(uriBTC, &resBTC)
	if b.bestHeightBTC == resBTC.BlockBook.BestHeight && resBTC.BlockBook.InSync {
		b.mu.Lock()
		b.stuckCountBTC++
		b.mu.Unlock()
	} else {
		b.mu.Lock()
		b.stuckCountBTC = 0
		b.mu.Unlock()
	}
	b.mu.Lock()
	b.isSyncedBTC = resBTC.BlockBook.InSync
	b.bestHeightBTC = resBTC.BlockBook.BestHeight
	if resBTC.BlockBook.BestHeight != 0 && resBTC.Backend.Blocks != 0 {
		b.syncBTCRatio = 100 * float64(resBTC.BlockBook.BestHeight) / float64(resBTC.Backend.Blocks)
	}
	if resBTC.BlockBook.MempoolSize != 0 && resBTC.BlockBook.InSyncMempool {
		b.isSyncedMempoolBTC = true
	}
	b.mu.Unlock()

	uriETH := fmt.Sprintf("http://%s/api/", b.nConf.BlockBookETH)
	b.api.GetRequest(uriETH, &resETH)

	if b.bestHeightETH == resETH.BlockBook.BestHeight && resETH.BlockBook.InSync {
		b.mu.Lock()
		b.stuckCountETH++
		b.mu.Unlock()
	} else {
		b.mu.Lock()
		b.stuckCountETH = 0
		b.mu.Unlock()
	}
	b.mu.Lock()
	b.isSyncedETH = resETH.BlockBook.InSync
	b.bestHeightETH = resETH.BlockBook.BestHeight
	if resETH.BlockBook.BestHeight != 0 && resETH.Backend.Blocks != 0 {
		b.syncETHRatio = 100 * float64(resETH.BlockBook.BestHeight) / float64(resETH.Backend.Blocks)
	}
	if resETH.BlockBook.MempoolSize != 0 && resETH.BlockBook.InSyncMempool {
		b.isSyncedMempoolETH = true
	}
	b.mu.Unlock()

	b.mu.Lock()
	log.Infof("BTC blockbook stuck_count: %d, ETH blockbook stuck_count: %d", b.stuckCountBTC, b.stuckCountETH)
	if b.stuckCountBTC >= 50 || b.stuckCountETH >= 50 {
		b.stuckCountBTC = 0
		b.stuckCountETH = 0
		log.Info("Restarting blockbook...")
		b.restartBlockbook()
	}
	b.mu.Unlock()
}

func (b *Bot) restartBlockbook() {
	extVars := map[string]string{
		"HOST_USER": b.hostUser,
	}
	path := fmt.Sprintf("./playbooks/mainnet_blockbook.yml")
	onSuccess := func() {
		log.Info("Blockbooks are restarted")
	}
	onError := func(err error) {
		log.Error(err)
	}
	b.execAnsible(path, extVars, onSuccess, onError)
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

func getDirSizeFromFile() (int, error) {
	path := fmt.Sprintf("/tmp/dir_size")
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	strs := strings.Split(string(str), "\t")
	log.Info(strs)
	intNum, _ := strconv.Atoi(strs[0])
	return intNum, nil
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
