package bot

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/SwingbyProtocol/node-installer/keystore"
	"github.com/binance-chain/go-sdk/common/types"
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
listen = "0.0.0.0"
port = 12121

[rest]
listen = "0.0.0.0"
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
testnet = "**is_testnet**"
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
rest_uri = "http://**btc_blockbook_endpoint**"
ws_uri = "ws://**btc_blockbook_endpoint**/websocket"
reward_addr = "**reward_address_btc**"
fixed_out_fee = 30000

[eth]
rest_uri = "http://**eth_blockbook_endpoint**"
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

func (b *Bot) generateKeys(basePath string) (bool, error) {
	pDataDirName := fmt.Sprintf("%s/data", basePath)
	pKeystoreFileName := fmt.Sprintf("%s/keystore.json", pDataDirName)
	stakeKeyPath := fmt.Sprintf("%s/key_%s.json", basePath, b.nConf.Network)
	_ = os.MkdirAll(pDataDirName, os.ModePerm)
	if _, _, err := keystore.LoadOrGenerate(pKeystoreFileName); err != nil {
		return false, err
	}
	pKeystore, err := keystore.ReadFromHome(pKeystoreFileName)
	if err != nil {
		return false, err
	}
	pP2PPubKey := pKeystore.P2pData.SK.Public()
	pP2PKeyHex := hex.EncodeToString(pP2PPubKey[:])

	addr, err := loadStakeKey(stakeKeyPath)
	if err == nil {
		b.nConf.StakeAddr = addr
		b.nConf.Memo = fmt.Sprintf("%s,%s", pP2PKeyHex, b.nConf.RewardAddressETH)
		return true, nil
	}

	pEntropy, err := bip39.NewEntropy(256)
	if err != nil {
		return false, err
	}
	// Gen a new address from new entropy
	pMnemonic, err := bip39.NewMnemonic(pEntropy)
	if err != nil {
		return false, err
	}
	if b.nConf.IsTestnet {
		types.Network = types.TestNetwork
	}
	pKey, err := keys.NewMnemonicKeyManager(pMnemonic)
	if err != nil {
		return false, err
	}
	log.Info(pMnemonic)
	b.nConf.StakeAddr = pKey.GetAddr().String()
	password, _ := generateRandomBytes(24)
	log.Infof("Deposit address: %s, pass: %s", b.nConf.StakeAddr, hex.EncodeToString(password))
	keydata, err := pKey.ExportAsKeyStore(hex.EncodeToString(password))
	if err != nil {
		return false, err
	}
	data, _ := json.Marshal(keydata)
	err = ioutil.WriteFile(stakeKeyPath, data, 0660)
	if err != nil {
		return false, err
	}
	// Check keystore
	_, err = keys.NewKeyStoreKeyManager(stakeKeyPath, hex.EncodeToString(password))
	if err != nil {
		return false, err
	}
	b.nConf.Memo = fmt.Sprintf("%s,%s", pP2PKeyHex, b.nConf.RewardAddressETH)
	return false, nil
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

func loadStakeKey(path string) (string, error) {
	data := keys.EncryptedKeyJSON{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.New("stake key is not exist")
	}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return "", err
	}
	return data.Address, nil
}
