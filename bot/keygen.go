package bot

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/SwingbyProtocol/node-installer/keystore"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/cosmos/go-bip39"
	log "github.com/sirupsen/logrus"
)

var bnbSeedNodes = []string{
	"tcp://data-seed-pre-0-s3.binance.org:80",
	"tcp://data-seed-pre-1-s3.binance.org:80",
	"tcp://data-seed-pre-0-s1.binance.org:80",
	"tcp://data-seed-pre-1-s1.binance.org:80",
	"tcp://data-seed-pre-2-s1.binance.org:80",
}

const baseConfig = `
[p2p]
moniker = "**node_moniker_placeholder**"
listen = "127.0.0.1"
port = 12121

[rest]
listen = "127.0.0.1"
port = 8067
tls_enabled = false

[logger]
level = "INFO"
max_file_size_MB = 100
max_backup_files = 100
max_retain_files_days = 28
use_console_logger = true
use_file_logger = true
compress = true

[swaps]
testnet = true
coin_1 = "BTCE"
coin_2 = "BTC"
stake_coin = "SWINGBY-888"

# stake_amount must be boosted by 10^8 for BNB Chain (this is 100,000)
stake_amount = 300000000000000
stake_time_secs = 300
fee_percent = 0.1

[tss]
participants = **participants_placeholder**
threshold = **threshold_placeholder**
keygen_until = "2020-07-23T12:00:00Z"

[btc]
rest_uri = "http://51.15.143.55:9130"
ws_uri = "ws://51.15.143.55:9130/websocket"
reward_addr = "2N8hwP1WmJrFF5QWABn38y63uYLhnJYJYTF"
fixed_out_fee = 30000

[eth]
rest_uri = "http://51.15.143.55:9131"
ws_uri = "ws://51.15.143.55:9131/websocket"
wallet_contract_addr = "0xebbc1dad17a79fb0bba3152497389ac552c1c24f"

[bnb]
rpc_uri = "**rpc_uri_placeholder**"
http_uri = "https://testnet-explorer.binance.org"
fixed_out_fee = 500`

func (b *Bot) generateConfig(path string) string {
	pDirName := fmt.Sprintf("%s/config", path)
	pDataDirName := fmt.Sprintf("%s/data", pDirName)
	pKeystoreFileName := fmt.Sprintf("%s/keystore.json", pDataDirName)
	_ = os.MkdirAll(pDataDirName, os.ModePerm)
	if err := keystore.GenerateInHome(pKeystoreFileName); err != nil {
		return ""
	}
	pKeystore, err := keystore.ReadFromHome(pKeystoreFileName)
	if err != nil {
		return ""
	}
	pP2PPubKey := pKeystore.P2pData.SK.Public()
	pP2PKeyHex := hex.EncodeToString(pP2PPubKey[:])

	// make the address for this staker
	pEntropy, err := bip39.NewEntropy(256)
	if err != nil {
		return ""
	}

	// generate the mnemonic/address for this peer
	pMnemonic, err := bip39.NewMnemonic(pEntropy)
	if err != nil {
		return ""
	}
	log.Infof("new mnemonic: %s", pMnemonic)
	pKey, err := keys.NewMnemonicKeyManager(pMnemonic)
	if err != nil {
		return ""
	}
	pAddr := pKey.GetAddr()
	log.Infof("desposit address: %s", pAddr)
	return fmt.Sprintf("%s,%s", pP2PKeyHex, pAddr.String())
}

func storeConfig(path string, moniker string, address string, threshold int, members int) {
	pDirName := fmt.Sprintf("%s/config", path)
	pConfigFileName := fmt.Sprintf("%s/config.toml", pDirName)
	stakeTxItem := fmt.Sprintf(`stake_tx = "%s"`, "user putin")
	stakeAddrItem := fmt.Sprintf(`stake_addr = "%s"`, address)
	rewardAddrItem := fmt.Sprintf(`reward_addr = "%s"`, address)
	newBaseConfig := strings.ReplaceAll(baseConfig, "**node_moniker_placeholder**", fmt.Sprintf("%s", moniker))
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**rpc_uri_placeholder**", bnbSeedNodes[0])
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**threshold_placeholder**", fmt.Sprintf("%d", threshold))
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**participants_placeholder**", fmt.Sprintf("%d", members))
	newConfigToml := fmt.Sprintf("%s\n%s\n%s\n%s\n", newBaseConfig, stakeTxItem, stakeAddrItem, rewardAddrItem)
	if err := ioutil.WriteFile(pConfigFileName, []byte(newConfigToml), os.ModePerm); err != nil {
		return
	}
}
