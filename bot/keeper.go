package bot

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func (b *Bot) checkBlockBook(coin string) {
	uri := fmt.Sprintf("%s/api/", b.nConf.BlockBookBTC)
	if coin == "ETH" {
		uri = fmt.Sprintf("%s/api/", b.nConf.BlockBookETH)
	}
	res := BlockBook{}
	err := b.api.GetRequest(uri, &res)
	b.mu.Lock()
	if b.nConf.BlockBookBTC != BlockBookBTC {
		b.infura = "global"
	} else {
		b.infura = "local"
	}
	if err != nil {
		b.stuckCount[coin]++
		b.bestHeight[coin] = 0
		b.SyncRatio[coin] = 0
		b.infuraVersions[coin] = ""
		b.mu.Unlock()
		return
	}
	b.stuckCount[coin]++

	if !b.isStartBB {
		b.isStartBB = true
	}

	if res.Backend.Version != "" {
		b.infuraVersions[coin] = res.Backend.Version
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
		} else {
			b.SyncRatio[coin] = 0
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

	b.mu.Lock()
	switch b.nConf.Network {
	case Network1:
		if regexp.MustCompile(BTCLockVersion).MatchString(b.infuraVersions["BTC"]) &&
			regexp.MustCompile(GethLockVersion).MatchString(b.infuraVersions["ETH"]) {
			b.validInfura = true
		} else {
			b.validInfura = false
		}
	case Network2:
		if regexp.MustCompile(BTCLockVersion).MatchString(b.infuraVersions["BTC"]) &&
			regexp.MustCompile(BSCLockVersion).MatchString(b.infuraVersions["ETH"]) {
			b.validInfura = true
		} else {
			b.validInfura = false
		}
	default:
		b.validInfura = false
	}

	b.mu.Unlock()

	b.mu.RLock()
	if !b.isStartBB {
		b.mu.RUnlock()
		return
	}
	if b.stuckCount["BTC"]%10 == 1 || b.stuckCount["ETH"]%10 == 1 {
		log.Infof("Blockbooks are not online stuck_count: BTC:%d, ETH:%d", b.stuckCount["BTC"], b.stuckCount["ETH"])
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
	path := fmt.Sprintf("./playbooks/%s/restart_blockbook.yml", b.nConf.Network)
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
