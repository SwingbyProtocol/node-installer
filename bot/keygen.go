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
coin_1 = "**coin_A**"
coin_2 = "**coin_B**"
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
rest_uri = "**btc_blockbook_endpoint**"
ws_uri = "ws://**btc_blockbook_endpoint**/websocket"
reward_addr = "**reward_address_btc**"
fixed_out_fee = 30000

[eth]
rest_uri = "**eth_blockbook_endpoint**"
ws_uri = "ws://**eth_blockbook_endpoint**/websocket"
wallet_contract_addr = "0xebbc1dad17a79fb0bba3152497389ac552c1c24f"
reward_addr = "**reward_address_eth**"

[bnb]
rpc_uri = "**rpc_uri_placeholder**"
http_uri = "https://testnet-explorer.binance.org"
fixed_out_fee = 500
stake_tx = "**stake_tx**"
stake_addr = "**stake_addr**"
reward_addr = "**reward_addr_bnb**"
`

func generateKeys(path string, rewardAddress string) (string, string) {
	pDirName := fmt.Sprintf("%s/config", path)
	pDataDirName := fmt.Sprintf("%s/data", pDirName)
	pKeystoreFileName := fmt.Sprintf("%s/keystore.json", pDataDirName)
	_ = os.MkdirAll(pDataDirName, os.ModePerm)
	if err := keystore.GenerateInHome(pKeystoreFileName); err != nil {
		return "", ""
	}
	pKeystore, err := keystore.ReadFromHome(pKeystoreFileName)
	if err != nil {
		return "", ""
	}
	pP2PPubKey := pKeystore.P2pData.SK.Public()
	pP2PKeyHex := hex.EncodeToString(pP2PPubKey[:])

	// make the address for this staker
	pEntropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", ""
	}

	// generate the mnemonic/address for this peer
	pMnemonic, err := bip39.NewMnemonic(pEntropy)
	if err != nil {
		return "", ""
	}
	log.Infof("new mnemonic: %s", pMnemonic)
	pKey, err := keys.NewMnemonicKeyManager(pMnemonic)
	if err != nil {
		return "", ""
	}
	pAddr := pKey.GetAddr()
	log.Infof("desposit address: %s", pAddr)
	return pAddr.String(), fmt.Sprintf("%s,%s", pP2PKeyHex, rewardAddress)
}

func storeConfig(path string, moniker string, threshold int, members int, coinA string, coinB string, blockBookBTCEndpoint string, blockBookETHndpoint string, addressBTC string, addressETH string, addressBNB string, stakeTx string) {
	pDirName := fmt.Sprintf("%s/config", path)
	pConfigFileName := fmt.Sprintf("%s/config.toml", pDirName)
	log.Info(pConfigFileName)
	newBaseConfig := strings.ReplaceAll(baseConfig, "**node_moniker_placeholder**", moniker)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_A**", coinA)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_B**", coinB)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**threshold_placeholder**", fmt.Sprintf("%d", threshold))
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**participants_placeholder**", fmt.Sprintf("%d", members))

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**btc_blockbook_endpoint**", blockBookBTCEndpoint)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_address_btc**", addressBTC)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**eth_blockbook_endpoint**", blockBookETHndpoint)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_address_eth**", addressETH)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**rpc_uri_placeholder**", bnbSeedNodes[0])
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_tx**", stakeTx)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_addr**", addressBNB)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_addr_bnb**", addressBNB)

	newConfigToml := fmt.Sprintf("%s\n", newBaseConfig)
	if err := ioutil.WriteFile(pConfigFileName, []byte(newConfigToml), os.ModePerm); err != nil {
		return
	}
}
