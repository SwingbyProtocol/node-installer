package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	commands := strings.Split(msg.Text, "@")
	cmd := commands[0]

	if cmd == "/start" {
		b.sayHello(msg.Chat.ID)
		return
	}
	if !b.validateChat(msg.Chat.ID) {
		return
	}

	b.handleReplyMessage(msg)

	b.handleSetupServer(cmd)
	b.handleSetupDomain(cmd)
	b.handleSetupYourBot(cmd)
	b.handleSetupNode(cmd)
	b.handleUpgradeYourBot(cmd)
	b.handleSetupInfura(cmd)
	b.handleDeployInfura(cmd)
	b.handleCheckStatus(cmd)
	b.handleDeployNode(cmd)
	b.handleDeployNodeDebug(cmd)
	b.handleEnableDomain(cmd)

	// Default response of say hi
	if cmd == "hi" || cmd == "Hi" {
		b.SendMsg(b.ID, `Let's start with /start`, false, false)
	}
}

func (b *Bot) sayHello(chatID int64) {
	if b.ID == 0 {
		b.ID = chatID
		b.SendMsg(b.ID, b.makeHelloText(), false, false)
		return
	}
	if !b.validateChat(chatID) {
		return
	}
	b.SendMsg(b.ID, b.makeHelloText(), false, false)
}

func (b *Bot) handleReplyMessage(msg *tgbotapi.Message) {
	// Handle for reply messages
	if msg.ReplyToMessage != nil {
		replyMsg := msg.Text
		prevMsg := msg.ReplyToMessage
		mode := b.Messages[prevMsg.MessageID]
		if mode == "setup_node_set_network" {
			b.updateNetwork(replyMsg)
			return
		}

		if mode == "setup_node_moniker" {
			b.updateNodeMoniker(replyMsg)
		}
		if mode == "setup_node_eth_addr" {
			b.updateETHAddr(replyMsg)
			return
		}
		if mode == "setup_node_stake_addr" {
			b.updateStakeAddr(replyMsg)
			return
		}
		// Set server configs
		if mode == "setup_ip_addr" {
			b.setupIPAddr(replyMsg)
			return
		}
		if mode == "setup_domain" {
			b.setupDomain(replyMsg)
			return
		}
		if mode == "setup_username" {
			b.setupUser(replyMsg)
			return
		}
	}
}

func (b *Bot) handleSetupServer(cmd string) {
	if cmd == "/setup_server_config" {
		// Disable if remote is `true`
		if b.isRemote {
			return
		}
		msg, err := b.SendMsg(b.ID, b.makeSetupIPText(), true, false)
		if err != nil {
			return
		}
		b.Messages[msg.MessageID] = "setup_ip_addr"
		return
	}
}

func (b *Bot) handleSetupDomain(cmd string) {
	if cmd == "/setup_domain" {
		newMsg, err := b.SendMsg(b.ID, b.setupDomainText(), true, false)
		if err != nil {
			return
		}
		b.Messages[newMsg.MessageID] = "setup_domain"
		return
	}
}

func (b *Bot) handleSetupYourBot(cmd string) {
	if cmd == "/setup_your_bot" {
		// Disable if remote is `true`
		if b.isRemote {
			return
		}
		err := b.loadHostAndKeys()
		if err != nil {
			log.Info(err)
			return
		}
		if b.checkProcess() {
			return
		}
		b.SendMsg(b.ID, makeDeployBotMessage(), false, false)
		extVars := map[string]string{
			"TAG":       b.botVersion,
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
		return
	}
}

func (b *Bot) handleSetupNode(cmd string) {
	if cmd == "/setup_node" {
		msg, err := b.SendMsg(b.ID, b.makeNodeText(), true, false)
		if err != nil {
			return
		}
		b.Messages[msg.MessageID] = "setup_node_set_network"
		return
	}
}

func (b *Bot) handleUpgradeYourBot(cmd string) {
	if cmd == "/upgrade_your_bot" {
		if !b.isRemote {
			return
		}
		err := b.loadHostAndKeys()
		if err != nil {
			log.Info(err)
			return
		}
		if b.checkProcess() {
			return
		}
		b.SendMsg(b.ID, makeUpgradeBotMessage(), false, false)
		contName := b.containerName
		if contName == "node_installer" {
			contName = "node_installer_fork"
		} else {
			contName = "node_installer"
		}
		extVars := map[string]string{
			"TAG":       b.botVersion,
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
		return
	}
}

func (b *Bot) handleSetupInfura(cmd string) {
	if cmd == "/setup_infura" {
		if b.checkProcess() {
			return
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
			return
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
		return
	}
}

func (b *Bot) handleDeployInfura(cmd string) {
	if cmd == "/deploy_infura" {
		if b.checkProcess() {
			return
		}
		if b.syncProgress < 99.99 {
			b.SendMsg(b.ID, rejectDeployInfuraMessage(), false, false)
			b.cooldown()
			return
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
			return
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
		return
	}
}

func (b *Bot) handleCheckStatus(cmd string) {
	if cmd == "/check_status" {
		if b.checkProcess() {
			return
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
		return
	}
}

func (b *Bot) handleDeployNode(cmd string) {
	if cmd == "/deploy_node" {
		if b.checkProcess() {
			return
		}
		if b.syncProgress <= 99.99 {
			b.SendMsg(b.ID, rejectDeployNodeMessage(), false, false)
			b.cooldown()
			return
		}
		extVars := map[string]string{
			"HOST_USER":        b.hostUser,
			"TAG":              b.nodeVersion,
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
		return
	}
}

func (b *Bot) handleDeployNodeDebug(cmd string) {
	if cmd == "/deploy_node_debug" {
		if b.checkProcess() {
			return
		}
		if b.syncProgress <= 99.99 {
			b.SendMsg(b.ID, rejectDeployNodeMessage(), false, false)
			b.cooldown()
			return
		}
		extVars := map[string]string{
			"HOST_USER":        b.hostUser,
			"TAG":              b.nodeVersion,
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
		return
	}
}

func (b *Bot) handleEnableDomain(cmd string) {
	if cmd == "/enable_domain" {
		if b.checkProcess() {
			return
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
		return
	}
}
