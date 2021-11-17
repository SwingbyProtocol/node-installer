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
		log.Info("Msg rejected -- worng chatID")
		return
	}

	b.handleReplyMessage(msg)

	b.handleSetupServer(cmd)
	b.handleSetupYourBot(cmd)
	b.handleSetupNode(cmd)
	b.handleShowMemo(cmd)

	b.handleSetupDomain(cmd)
	b.handleEnableDomain(cmd)
	b.handleStopNginx(cmd)

	b.handleDeployNode(cmd)
	b.handleDeployNodeDebug(cmd)
	b.handleStopNode(cmd)

	b.handleGetLogs(cmd)
	b.handleStakeUpdateNode(cmd)

	b.handleSetupInfura(cmd)
	b.handleResyncInfura(cmd)
	b.handleRemoveInfura(cmd)
	b.handleDeployInfura(cmd)
	b.handleResetGeth(cmd)
	b.handleSetGlobalInfura(cmd)
	b.handleSetLocalInfura(cmd)

	b.handleCheckStatus(cmd)
	b.handleUpgradeBot(cmd)

	b.handleOpenGethPort(cmd)
	b.handleOpenBlockBooksPort(cmd)
	// Default response of say hi
	if cmd == "/hi" || cmd == "/Hi" || cmd == "/help" {
		b.SendMsg(b.ID, `OK. Let's start with /start`, false, false)
	}
}

func (b *Bot) sayHello(chatID int64) {
	if b.ID == 0 {
		b.ID = chatID
		b.SendMsg(b.ID, b.makeHelloText(), false, false)
		b.SendMsg(b.ID, b.instructionMention(), false, false)
		return
	}
	if !b.validateChat(chatID) {
		log.Info("Msg rejected -- worng chatID")
		return
	}
	b.SendMsg(b.ID, b.makeHelloText(), false, false)
}

func (b *Bot) validateChat(chatID int64) bool {
	if b.ID == chatID {
		return true
	}
	return false
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
			return
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

func (b *Bot) handleSetupYourBot(cmd string) {
	if cmd == "/setup_your_bot" {
		// Disable if remote is `true`
		if b.isRemote {
			return
		}
		err := b.loadAndCreateHostIPAndKeys()
		if err != nil {
			log.Info(err)
			return
		}
		if b.checkProcess(true) {
			return
		}
		b.SendMsg(b.ID, makeDeployBotMessage(), false, false)
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
		}
		onSuccess := func() {
			diskSpace, _ := getDiskSpaceFromFile()
			if diskSpace <= minimumMountPathSizeMiB[Network2] {
				b.SendMsg(b.ID, rejectDeployInfuraByDiskSpaceMessage(), false, false)
				b.cooldown()
				return
			}
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
				b.SendMsg(b.ID, errorDeployBotMessage(), false, false)
				b.cooldown()
			}
			b.execAnsible("./playbooks/bot_install.yml", extVars, onSuccess, onError)
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorDeployBotMessage(), false, false)
			b.cooldown()
		}
		b.execAnsible("./playbooks/setup_node.yml", extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleSetupNode(cmd string) {
	if cmd == "/setup_node" {
		if !b.isRemote {
			return
		}
		msg, err := b.SendMsg(b.ID, b.makeNodeText(), true, false)
		if err != nil {
			return
		}
		b.Messages[msg.MessageID] = "setup_node_set_network"
		return
	}
}

func (b *Bot) handleSetupDomain(cmd string) {
	if cmd == "/setup_domain" {
		if !b.isRemote {
			return
		}
		newMsg, err := b.SendMsg(b.ID, b.setupDomainText(), true, false)
		if err != nil {
			return
		}
		b.Messages[newMsg.MessageID] = "setup_domain"
		return
	}
}

func (b *Bot) handleShowMemo(cmd string) {
	if cmd == "/show_timelock_memo" {
		if !b.isRemote {
			return
		}
		_, err := b.SendMsg(b.ID, b.showMemoText(b.nConf.Memo, b.nConf.StakeAddr), true, false)
		if err != nil {
			return
		}
		return
	}
}

func (b *Bot) handleUpgradeBot(cmd string) {
	if cmd == "/upgrade_bot" {
		if !b.isRemote {
			return
		}
		err := b.loadAndCreateHostIPAndKeys()
		if err != nil {
			log.Info(err)
			return
		}
		if b.checkProcess(true) {
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
			"TAG":       b.nextBotVersion,
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
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorDeployBotMessage(), false, false)
			b.cooldown()
		}
		b.execAnsible("./playbooks/bot_install.yml", extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleSetupInfura(cmd string) {
	if cmd == "/setup_infura" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
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
		onSuccess := func() {
			b.SendMsg(b.ID, doneSetupInfuraMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorSetupInfuraMessage(), false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/%s/snapshot_sync.yml", b.nConf.Network)
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleResyncInfura(cmd string) {
	if cmd == "/resync_infura" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		if !b.isConfirmed["resync_infura"] {
			b.SendMsg(b.ID, confirmResyncInfuraMessage(), false, false)
			b.mu.Lock()
			b.isConfirmed["resync_infura"] = true
			b.mu.Unlock()
			go func() {
				time.Sleep(10 * time.Second)
				b.mu.Lock()
				b.isConfirmed["resync_infura"] = false
				b.mu.Unlock()
			}()
			b.cooldown()
			return
		}
		b.mu.Lock()
		b.isConfirmed["resync_infura"] = false
		b.mu.Unlock()
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
		}
		b.SendMsg(b.ID, makeResyncInfuraMessage(), false, false)
		onSuccess := func() {
			b.SendMsg(b.ID, doneResyncInfuraMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorResyncInfuraMessage(), false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/%s/snapshot_resync.yml", b.nConf.Network)
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleResetGeth(cmd string) {
	if cmd == "/reset_geth" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		if !b.isConfirmed["reset_geth"] {
			b.SendMsg(b.ID, confirmResetGethMessage(), false, false)
			b.mu.Lock()
			b.isConfirmed["reset_geth"] = true
			b.mu.Unlock()
			go func() {
				time.Sleep(10 * time.Second)
				b.mu.Lock()
				b.isConfirmed["reset_geth"] = false
				b.mu.Unlock()
			}()
			b.cooldown()
			return
		}
		b.mu.Lock()
		b.isConfirmed["reset_geth"] = false
		b.mu.Unlock()
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
		}
		b.SendMsg(b.ID, makeResetGethMessage(), false, false)
		onSuccess := func() {
			b.SendMsg(b.ID, doneResetGethMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorResetGethMessage(), false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/%s/reset_geth.yml", b.nConf.Network)
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleRemoveInfura(cmd string) {
	if cmd == "/remove_infura" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		if !b.isConfirmed["remove_infura"] {
			b.SendMsg(b.ID, confirmRemoveInfuraMessage(), false, false)
			b.mu.Lock()
			b.isConfirmed["remove_infura"] = true
			b.mu.Unlock()
			go func() {
				time.Sleep(10 * time.Second)
				b.mu.Lock()
				b.isConfirmed["remove_infura"] = false
				b.mu.Unlock()
			}()
			b.cooldown()
			return
		}
		b.mu.Lock()
		b.isConfirmed["remove_infura"] = false
		b.mu.Unlock()
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
		}
		b.SendMsg(b.ID, makeRemoveInfuraMessage(), false, false)
		onSuccess := func() {
			b.SendMsg(b.ID, doneRemoveInfuraMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorRemoveInfuraMessage(), false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/%s/remove_infura.yml", b.nConf.Network)
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleDeployInfura(cmd string) {
	if cmd == "/deploy_infura" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		// if b.syncProgress < 99.99 {
		// 	b.SendMsg(b.ID, rejectDeployInfuraMessage(), false, false)
		// 	b.cooldown()
		// 	return
		// }
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
		onSuccess := func() {
			b.SendMsg(b.ID, doneDeployInfuraMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorDeployInfuraMessage(), false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/%s/deploy_infura.yml", b.nConf.Network)
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleSetGlobalInfura(cmd string) {
	if cmd == "/set_global_infura" {
		if !b.isRemote {
			return
		}
		b.nConf.SetGlobalNode()
		b.nConf.saveConfig()
		b.SendMsg(b.ID, "Ok. set global mode", false, false)
		return
	}
}

func (b *Bot) handleSetLocalInfura(cmd string) {
	if cmd == "/set_local_infura" {
		if !b.isRemote {
			return
		}
		b.nConf.SetLocalNode()
		b.nConf.saveConfig()
		b.SendMsg(b.ID, "Ok. set local mode", false, false)
		return
	}
}

func (b *Bot) handleCheckStatus(cmd string) {
	if cmd == "/check_status" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
			"IP_ADDR":   b.nodeIP,
		}
		b.SendMsg(b.ID, b.makeCheckNodeMessage(), false, false)
		onSuccess := func() {
			syncedSize, _ := getDirSizeFromFile()
			availableSize, _ := getAvailableDiskSpaceFromFile()
			parcent := 100.00 * float64(syncedSize) / float64(syncSnapshotBytes[b.nConf.Network])
			if parcent >= 99.998 {
				b.syncProgress = 99.99
			}
			if parcent < 99.99 {
				b.syncProgress = parcent
			}
			if b.SyncRatio["BTC"] == 100.00 && b.SyncRatio["ETH"] == 100.00 {
				b.syncProgress = 100.00
			}
			b.SendMsg(b.ID, b.checkNodeMessage(availableSize), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorCheckNodeMessage(), false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/check_status.yml")
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) autoCheckSpace() {
	if !b.isRemote {
		return
	}
	if b.checkProcess(false) {
		return
	}
	extVars := map[string]string{
		"HOST_USER": b.hostUser,
		"IP_ADDR":   b.nodeIP,
	}
	onSuccess := func() {
		availableSize, _ := getAvailableDiskSpaceFromFile()
		availableGBs := availableSize / 1024
		if availableGBs < 50 {
			b.SendMsg(b.ID, informStorageIssue(), false, false)
		}
		b.cooldown()
	}
	onError := func(err error) {
		log.Info(err)
		b.cooldown()
	}
	path := fmt.Sprintf("./playbooks/check_status.yml")
	b.execAnsible(path, extVars, onSuccess, onError)
	return
}

func (b *Bot) handleOpenGethPort(cmd string) {
	if cmd == "/open_geth_port" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
			"IP_ADDR":   b.nodeIP,
		}
		port := "8545"
		if b.nConf.Network == Network2 {
			port = "8575"
		}
		//b.SendMsg(b.ID, makeCheckNodeMessage(), false, false)
		onSuccess := func() {
			//b.SendMsg(b.ID, b.checkNodeMessage(), false, false)
			b.SendMsg(b.ID, fmt.Sprintf("Ok. port %s is opened", port), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			//b.SendMsg(b.ID, errorCheckNodeMessage(), false, false)
			b.SendMsg(b.ID, "Something wrong", false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/%s/open_geth_port.yml", b.nConf.Network)
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleOpenBlockBooksPort(cmd string) {
	if cmd == "/open_blockbooks_port" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
			"IP_ADDR":   b.nodeIP,
		}
		port := "9130, 9131"
		if b.nConf.Network == Network2 {
			port = "9130, 9132"
		}
		//b.SendMsg(b.ID, makeCheckNodeMessage(), false, false)
		onSuccess := func() {
			//b.SendMsg(b.ID, b.checkNodeMessage(), false, false)
			b.SendMsg(b.ID, fmt.Sprintf("Ok. port %s is opened", port), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			//b.SendMsg(b.ID, errorCheckNodeMessage(), false, false)
			b.SendMsg(b.ID, "Something wrong", false, false)
			b.cooldown()
		}
		path := fmt.Sprintf("./playbooks/%s/open_blockbooks_port.yml", b.nConf.Network)
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleDeployNode(cmd string) {
	if cmd == "/deploy_node" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		if b.syncProgress <= 99.99 {
			b.SendMsg(b.ID, rejectDeployNodeByInfuraMessage(), false, false)
			b.cooldown()
			return
		}
		if !b.validInfura {
			b.SendMsg(b.ID, rejectDeployNodeByUpgradeInfuraMessage(), false, false)
			b.cooldown()
			return
		}
		if b.nConf.checkConfig() != nil {
			b.SendMsg(b.ID, rejectDeployNodeByConfigMessage(), false, false)
			b.cooldown()
			return
		}
		b.nConf.storeConfigToml()
		b.nConf.saveConfig()
		extVars := map[string]string{
			"HOST_USER":        b.hostUser,
			"TAG":              b.nodeVersion,
			"NETWORK":          b.nConf.Network,
			"BOOTSTRAP_NODE_1": b.nConf.BootstrapNode[0],
			"BOOTSTRAP_NODE_2": b.nConf.BootstrapNode[1],
			"BOOTSTRAP_NODE_3": b.nConf.BootstrapNode[2],
			"K_UNTIL":          b.nConf.KeygenUntil,
			"LOG_LEVEL":        "INFO",
		}
		b.SendMsg(b.ID, b.makeDeployNodeMessage(), false, false)
		path := fmt.Sprintf("./playbooks/%s/deploy_node.yml", b.nConf.Network)
		onSuccess := func() {
			b.SendMsg(b.ID, doneDeployNodeMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorDeployNodeMessage(), false, false)
			b.cooldown()
		}
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleDeployNodeDebug(cmd string) {
	if cmd == "/deploy_node_debug" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		if b.syncProgress <= 99.99 {
			b.SendMsg(b.ID, rejectDeployNodeByInfuraMessage(), false, false)
			b.cooldown()
			return
		}
		if !b.validInfura {
			b.SendMsg(b.ID, rejectDeployNodeByUpgradeInfuraMessage(), false, false)
			b.cooldown()
			return
		}
		if b.nConf.checkConfig() != nil {
			b.SendMsg(b.ID, rejectDeployNodeByConfigMessage(), false, false)
			b.cooldown()
			return
		}
		extVars := map[string]string{
			"HOST_USER":        b.hostUser,
			"TAG":              b.nodeVersion,
			"NETWORK":          b.nConf.Network,
			"BOOTSTRAP_NODE_1": b.nConf.BootstrapNode[0],
			"BOOTSTRAP_NODE_2": b.nConf.BootstrapNode[1],
			"BOOTSTRAP_NODE_3": b.nConf.BootstrapNode[2],
			"K_UNTIL":          b.nConf.KeygenUntil,
			"LOG_LEVEL":        "DEBUG",
		}
		b.SendMsg(b.ID, b.makeDeployNodeMessage(), false, false)
		path := fmt.Sprintf("./playbooks/%s/deploy_node.yml", b.nConf.Network)
		onSuccess := func() {
			b.SendMsg(b.ID, doneDeployNodeMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorDeployNodeMessage(), false, false)
			b.cooldown()
		}
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleStopNginx(cmd string) {
	if cmd == "/stop_nginx" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
		}
		b.SendMsg(b.ID, b.makeStopNginxMessage(), false, false)
		path := fmt.Sprintf("./playbooks/stop_nginx.yml")
		onSuccess := func() {
			b.SendMsg(b.ID, b.doneStopNginxMessage(), false, false)
			b.cooldown()

		}
		onError := func(err error) {
			b.SendMsg(b.ID, b.errorStopNginxMessage(), false, false)
			b.cooldown()
		}
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleStopNode(cmd string) {
	if cmd == "/stop_node" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
		}
		b.SendMsg(b.ID, b.makeStopNodeMessage(), false, false)
		path := fmt.Sprintf("./playbooks/stop_node.yml")
		onSuccess := func() {
			b.SendMsg(b.ID, b.doneStopNodeMessage(), false, false)
			b.cooldown()

		}
		onError := func(err error) {
			b.SendMsg(b.ID, b.errorStopNodeMessage(), false, false)
			b.cooldown()
		}
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleStakeUpdateNode(cmd string) {
	if cmd == "/stake_update" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		extVars := map[string]string{
			"HOST_USER": b.hostUser,
		}
		b.SendMsg(b.ID, b.makeUpdateStakeNodeMessage(), false, false)
		path := fmt.Sprintf("./playbooks/stop_node.yml")
		onSuccess := func() {
			b.SendMsg(b.ID, b.doneUpdateStakeNodeMessage(), false, false)
			b.cooldown()
		}
		onError := func(err error) {
			b.SendMsg(b.ID, b.errorUpdateStakeNodeMessage(), false, false)
			b.cooldown()
		}
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
}

func (b *Bot) handleGetLogs(cmd string) {
	if cmd == "/get_node_logs" {
		if !b.isRemote {
			return
		}
		path := fmt.Sprintf("%s/%s", DataPath, b.nConf.Network)
		err := b.sendLogFile(path)
		if err != nil {
			b.SendMsg(b.ID, errorLogFileMessage(), false, false)
		}
		return
	}
}

func (b *Bot) handleEnableDomain(cmd string) {
	if cmd == "/enable_domain" {
		if !b.isRemote {
			return
		}
		if b.checkProcess(true) {
			return
		}
		extVars := map[string]string{
			"HOST_USER":     b.hostUser,
			"DOMAIN":        b.nConf.Domain,
			"WITH_IDNEXERS": "false",
		}
		b.SendMsg(b.ID, b.makeDomainMessage(), false, false)
		path := fmt.Sprintf("./playbooks/enable_domain.yml")
		onSuccess := func() {
			b.SendMsg(b.ID, b.doneDomainMessage(), false, false)
			b.cooldown()

		}
		onError := func(err error) {
			b.SendMsg(b.ID, errorDomainMessage(), false, false)
			b.cooldown()

		}
		b.execAnsible(path, extVars, onSuccess, onError)
		return
	}
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
	err := b.loadAndCreateHostIPAndKeys()
	if err != nil {
		text := fmt.Sprintf("SSH_KEY loading error. please check data/ssh_key file again")
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
		b.nConf.storeConfigToml()
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
	b.nConf.SetLocalNode()
	b.nConf.storeConfigToml()
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
	path := fmt.Sprintf("%s/%s", DataPath, b.nConf.Network)
	_, err := b.generateKeys(path)
	if err != nil {
		log.Error(err)
		return
	}
	b.SendMsg(b.ID, b.makeStakeAddrText(), false, true)
	newMsg, _ := b.SendMsg(b.ID, b.askStakeAddrText(), true, false)
	b.Messages[newMsg.MessageID] = "setup_node_stake_addr"
}

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
		b.nConf.SetNetwork(network)
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
