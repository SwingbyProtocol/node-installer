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

const (
	nodeVersion = "0.1.0"
	botVersion  = "1.0.0"
	dataPath    = "./data"
	network1    = "mainnet_btc_eth"
	network2    = "mainnet_btc_bc"
	network3    = "testnet_tbtc_goerli"
	network4    = "testnet_tbtc_bc"
)

var Networks = map[string]string{
	"1": network1,
	"2": network2,
	"3": network3,
	"4": network4,
}

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
	syncRatio       map[string]float64
	syncProgress    float64
}

func NewBot(token string) (*Bot, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot := &Bot{
		Messages:        make(map[int]string),
		ID:              0,
		mu:              new(sync.RWMutex),
		bot:             b,
		api:             api.NewResolver("", 200),
		nodeVersion:     nodeVersion,
		botVersion:      botVersion,
		hostUser:        "root",
		containerName:   "node_installer",
		nConf:           NewNodeConfig(),
		isConfirmed:     make(map[string]bool),
		stuckCount:      make(map[string]int),
		isSynced:        make(map[string]bool),
		isSyncedMempool: make(map[string]bool),
		syncRatio:       make(map[string]float64),
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

	if b.isRemote {
		b.startBBKeeper()
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
		storeSSHKeyfile(os.Getenv("SSH_KEY"))
		log.Info("SSH private key is stored")
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
	uri := fmt.Sprintf("http://%s/api/", b.nConf.BlockBookBTC)
	if coin == "ETH" {
		uri = fmt.Sprintf("http://%s/api/", b.nConf.BlockBookETH)
	}
	res := BlockBook{}
	err := b.api.GetRequest(uri, &res)
	b.mu.Lock()
	if err != nil {
		b.stuckCount[coin]++
		b.mu.Unlock()
		return
	}
	b.stuckCount[coin]++
	if b.bestHeight[coin] != res.BlockBook.BestHeight && res.BlockBook.InSync {
		b.stuckCount[coin] = 0
		b.isSynced[coin] = res.BlockBook.InSync
		b.bestHeight[coin] = res.BlockBook.BestHeight
		if res.BlockBook.BestHeight != 0 && res.Backend.Blocks != 0 {
			b.syncRatio[coin] = 100 * float64(res.BlockBook.BestHeight) / float64(res.Backend.Blocks)
			if b.syncRatio[coin] == 100.00 {
				b.syncRatio[coin] = 99
			}
		}
		if res.BlockBook.MempoolSize != 0 && res.BlockBook.InSyncMempool {
			b.syncRatio[coin] = 100.00
			b.isSyncedMempool[coin] = true
		}
	}
	b.mu.Unlock()
}

func (b *Bot) checkBlockBooks() {
	b.checkBlockBook("BTC")
	b.checkBlockBook("ETH")
	b.mu.Lock()
	//log.Infof("BTC blockbook stuck_count: %d, ETH blockbook stuck_count: %d", b.stuckCountBTC, b.stuckCountETH)
	if b.stuckCount["BTC"] >= 70 || b.stuckCount["ETH"] >= 50 {
		b.stuckCount["BTC"] = 0
		b.stuckCount["ETH"] = 0
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
