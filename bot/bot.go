package bot

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/SwingbyProtocol/tx-indexer/api"
	ansibler "github.com/apenella/go-ansible"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	Messages        map[int]string
	ID              int64
	mu              *sync.RWMutex
	bot             *tgbotapi.BotAPI
	api             *api.Resolver
	nodeVersion     string
	botVersion      string
	hostUser        string
	nodeIP          string
	containerName   string
	sshKey          string
	nConf           *NodeConfig
	isRemote        bool
	isLocked        bool
	isConfirmed     map[string]bool
	stuckCount      map[string]int
	bestHeight      map[string]int
	isSynced        map[string]bool
	isSyncedMempool map[string]bool
	SyncRatio       map[string]float64
	syncProgress    float64
	isStartBB       bool
}

func NewBot(token string) (*Bot, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	ver, err := getVersion()
	if err != nil {
		panic(err)
	}
	bot := &Bot{
		Messages:        make(map[int]string),
		ID:              0,
		mu:              new(sync.RWMutex),
		bot:             b,
		api:             api.NewResolver("", 200),
		nodeVersion:     ver.NodeVersion,
		botVersion:      ver.BotVersion,
		hostUser:        "root",
		containerName:   "node_installer",
		nConf:           NewNodeConfig(),
		isConfirmed:     make(map[string]bool),
		stuckCount:      make(map[string]int),
		bestHeight:      make(map[string]int),
		isSynced:        make(map[string]bool),
		isSyncedMempool: make(map[string]bool),
		SyncRatio:       make(map[string]float64),
	}
	return bot, nil
}

func (b *Bot) Start() {
	b.bot.Debug = false
	b.api.SetTimeout(15 * time.Second)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	b.loadSystemEnv()
	b.loadAndCreateHostIPAndKeys()
	b.nConf.loadConfig()
	log.Infof("Loaded KeygenUntil: %s", b.nConf.KeygenUntil)
	log.Infof("Authorized on bot account: %s", b.bot.Self.UserName)
	log.Infof("Bot is ready with Version: %s, [node: %s]", b.botVersion, b.nodeVersion)

	if b.isRemote {
		b.startBBKeeper()
	}

	if b.nConf.BlockBookBTCWS == "" {
		b.nConf = NewNodeConfig()
	}

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
		b.handleMessage(update.Message)
	}
}

func (b *Bot) startBBKeeper() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		time.Sleep(10 * time.Second)
		b.checkBlockBooks()
		for {
			<-ticker.C
			b.checkBlockBooks()
			b.checkNewVersion()
		}
	}()
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

func (b *Bot) loadSystemEnv() {
	if os.Getenv("REMOTE") == "true" {
		b.isRemote = true
		log.Infof("Set Remote mode")
	}
	if os.Getenv("CHAT_ID") != "" {
		intID, err := strconv.Atoi(os.Getenv("CHAT_ID"))
		if err == nil {
			b.ID = int64(intID)
		}
		log.Infof("Set ChatID=%d", b.ID)
	}
	if os.Getenv("IP_ADDR") != "" {
		b.nodeIP = os.Getenv("IP_ADDR")
		log.Infof("Set IP_ADDR=%s", b.nodeIP)

	}
	if os.Getenv("CONT_NAME") != "" {
		b.containerName = os.Getenv("CONT_NAME")
		log.Infof("Set CONT_NAME=%s", b.containerName)
	}
	if os.Getenv("HOST_USER") != "" {
		b.hostUser = os.Getenv("HOST_USER")
		log.Infof("Set HOST_USER=%s", b.hostUser)
	}
	if os.Getenv("SSH_KEY") != "" {
		b.sshKey = os.Getenv(("SSH_KEY"))
		log.Info("SSH private key is stored")
	}
	if os.Getenv("TAG") != "" {
		b.botVersion = os.Getenv("TAG")
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

func (b *Bot) loadAndCreateHostIPAndKeys() error {
	host, err := getFileHostfile()
	if err != nil {
		generateHostsfile(b.nodeIP, "server")
		host = b.nodeIP
	}
	b.nodeIP = host
	log.Infof("Loaded Server IPv4:%s", host)
	// load ssh key file
	key, err := getFileSSHKeyfie()
	if err != nil {
		storeSSHKeyfile(b.sshKey)
		key = b.sshKey
	}
	err = storeSSHKeyfile(key)
	if err != nil {
		return err
	}
	updatedKey, err := getFileSSHKeyfie()
	if err != nil {
		return err
	}
	log.Infof("Loaded SSH priv key")
	b.sshKey = updatedKey
	return nil
}

func (b *Bot) execAnsible(playbookPath string, extVars map[string]string, onSuccess func(), onError func(err error)) {
	ansibler.AnsibleAvoidHostKeyChecking()

	sshKeyFilePath := fmt.Sprintf("%s/ssh_key", DataPath)
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
		//log.Info(playbook.String())
		err := playbook.Run()
		if err != nil {
			onError(err)
			return
		}
		onSuccess()
	}()
}

func (b *Bot) checkBlockBook(coin string) {
	uri := fmt.Sprintf("%s/api/", b.nConf.BlockBookBTC)
	if coin == "ETH" {
		uri = fmt.Sprintf("%s/api/", b.nConf.BlockBookETH)
	}
	res := BlockBook{}
	err := b.api.GetRequest(uri, &res)
	b.mu.Lock()
	if err != nil {
		b.stuckCount[coin]++
		b.bestHeight[coin] = 0
		b.SyncRatio[coin] = 0
		b.mu.Unlock()
		return
	}
	b.stuckCount[coin]++

	if !b.isStartBB {
		b.isStartBB = true
	}

	if b.bestHeight[coin] <= res.BlockBook.BestHeight {
		b.stuckCount[coin] = 0
		b.isSynced[coin] = res.BlockBook.InSync
		b.bestHeight[coin] = res.BlockBook.BestHeight
		if res.BlockBook.BestHeight != 0 && res.Backend.Blocks != 0 {
			syncRatio := 100 * float64(res.BlockBook.BestHeight) / float64(res.Backend.Blocks)
			if syncRatio >= 99.99 {
				b.SyncRatio[coin] = 99.99
			} else {
				b.SyncRatio[coin] = syncRatio
			}
		}
	}

	if b.SyncRatio[coin] == 99.99 && res.BlockBook.MempoolSize != 0 && res.BlockBook.InSyncMempool {
		b.SyncRatio[coin] = 100.00
		b.isSyncedMempool[coin] = true
	}
	b.mu.Unlock()
}

func (b *Bot) checkBlockBooks() {

	b.checkBlockBook("BTC")
	b.checkBlockBook("ETH")
	b.mu.RLock()
	if !b.isStartBB {
		b.mu.RUnlock()
		return
	}
	if b.stuckCount["BTC"]%10 == 1 || b.stuckCount["ETH"]%10 == 1 {
		log.Infof("Bblockbooks are not online stuck_count: BTC:%d, ETH:%d", b.stuckCount["BTC"], b.stuckCount["ETH"])
	}
	if b.stuckCount["BTC"] >= 71 || b.stuckCount["ETH"] >= 51 {
		b.mu.RUnlock()
		log.Info("Restarting blockbook...")
		b.restartBlockbooks()
		return
	}
	b.mu.RUnlock()
}

func (b *Bot) restartBlockbooks() {
	extVars := map[string]string{
		"HOST_USER": b.hostUser,
	}
	path := fmt.Sprintf("./playbooks/mainnet_blockbook.yml")
	onSuccess := func() {
		log.Info("Blockbooks are restarted")
		b.mu.Lock()
		b.stuckCount["BTC"] = 0
		b.stuckCount["ETH"] = 0
		b.mu.Unlock()
	}
	onError := func(err error) {
		log.Error(err)
	}
	b.execAnsible(path, extVars, onSuccess, onError)
}

func (b *Bot) checkNewVersion() {
	v := Version{}
	err := b.api.GetRequest(VersionJSON, &v)
	if err != nil {
		log.Info(err)
		return
	}
	bVersion, nVersion := b.Versions()
	if v.BotVersion != bVersion && v.NodeVersion != nVersion {
		log.Infof("the new version of bot [v%s] node [v%s] is coming!", v.BotVersion, v.NodeVersion)
		b.SendMsg(b.ID, upgradeBothMessage(v.BotVersion, v.NodeVersion), false, false)
		b.SetVersion(v.BotVersion, v.NodeVersion)
		return
	}
	if v.BotVersion != bVersion {
		log.Infof("the new version of bot [v%s] is coming!", v.BotVersion)
		b.SendMsg(b.ID, upgradeBotMessage(v.BotVersion), false, false)
		b.SetVersion(v.BotVersion, nVersion)
		return
	}
	if v.NodeVersion != nVersion {
		log.Infof("the new version of node [v%s] is coming!", v.NodeVersion)
		b.SendMsg(b.ID, upgradeNodeMessage(v.NodeVersion), false, false)
		b.SetVersion(bVersion, v.NodeVersion)
		return
	}
}

func (b *Bot) Versions() (string, string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.botVersion, b.nodeVersion
}

func (b *Bot) SetVersion(bV string, nV string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.botVersion = bV
	b.nodeVersion = nV
}
