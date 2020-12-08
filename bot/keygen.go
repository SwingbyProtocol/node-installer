package bot

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/SwingbyProtocol/node-installer/keystore"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/common/uuid"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/cosmos/go-bip39"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
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

type cipherParams struct {
	IV string `json:"iv"`
}

type CryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherParams           `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type EncryptedKey struct {
	Crypto  CryptoJSON `json:"crypto"`
	Id      string     `json:"id"`
	Version int        `json:"version"`
}

func (b *Bot) generateKeys(network string, rewardAddress string, isTestnet bool) (string, error) {
	pDataDirName := fmt.Sprintf("%s/data", network)
	pKeystoreFileName := fmt.Sprintf("%s/keystore.json", pDataDirName)
	_ = os.MkdirAll(pDataDirName, os.ModePerm)
	if err := keystore.GenerateInHome(pKeystoreFileName); err != nil {
		return "", err
	}
	pKeystore, err := keystore.ReadFromHome(pKeystoreFileName)
	if err != nil {
		return "", err
	}
	pP2PPubKey := pKeystore.P2pData.SK.Public()
	pP2PKeyHex := hex.EncodeToString(pP2PPubKey[:])

	// make the address for this staker
	pEntropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}

	// generate the mnemonic/address for this peer
	pMnemonic, err := bip39.NewMnemonic(pEntropy)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/data/mnemonic", network), []byte(pMnemonic), 0666)
	if err != nil {
		return "", err
	}
	if isTestnet {
		types.Network = types.TestNetwork
	}
	pKey, err := keys.NewMnemonicKeyManager(pMnemonic)
	if err != nil {
		return "", err
	}
	pAddr := pKey.GetAddr()
	b.stakeAddr = pAddr.String()
	log.Infof("desposit address: %s", pAddr)
	return fmt.Sprintf("%s,%s", pP2PKeyHex, rewardAddress), nil
}

func (b *Bot) storeConfig(network string, threshold int, members int) error {
	pConfigFileName := fmt.Sprintf("%s/config.toml", network)
	newBaseConfig := strings.ReplaceAll(baseConfig, "**node_moniker_placeholder**", b.moniker)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_A**", b.coinA)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_B**", b.coinB)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**is_testnet**", fmt.Sprintf("%t", b.isTestnet))

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**threshold_placeholder**", fmt.Sprintf("%d", threshold))
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**participants_placeholder**", fmt.Sprintf("%d", members))

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**btc_blockbook_endpoint**", b.blockBookBTC)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_address_btc**", b.rewardAddressBTC)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**eth_blockbook_endpoint**", b.blockBookETH)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_address_eth**", b.rewardAddressETH)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**rpc_uri_placeholder**", bnbSeedNodes[0])
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_tx**", b.stakeTx)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_addr**", b.stakeAddr)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_addr_bnb**", b.rewardAddressBNB)

	newConfigToml := fmt.Sprintf("%s\n", newBaseConfig)
	if err := ioutil.WriteFile(pConfigFileName, []byte(newConfigToml), os.ModePerm); err != nil {
		return err
	}
	return nil
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

func exportKeystore(path string, text string, password string) error {
	salt, err := generateRandomBytes(32)
	if err != nil {
		return err
	}
	iv, err := generateRandomBytes(16)
	if err != nil {
		return err
	}
	scryptParamsJSON := make(map[string]interface{}, 4)
	scryptParamsJSON["prf"] = "hmac-sha256"
	scryptParamsJSON["dklen"] = 32
	scryptParamsJSON["salt"] = hex.EncodeToString(salt)
	scryptParamsJSON["c"] = 262144

	cipherParamsJSON := cipherParams{IV: hex.EncodeToString(iv)}
	derivedKey := pbkdf2.Key([]byte(password), salt, 262144, 32, sha256.New)
	encryptKey := derivedKey[:32]
	cipherText, err := aesCTRXOR(encryptKey, []byte(text), iv)
	if err != nil {
		return err
	}

	hasher := sha3.NewLegacyKeccak512()
	_, err = hasher.Write(derivedKey[16:32])
	if err != nil {
		return err
	}
	_, err = hasher.Write(cipherText)
	if err != nil {
		return err
	}
	mac := hasher.Sum(nil)

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	cryptoStruct := CryptoJSON{
		Cipher:       "aes-256-ctr",
		CipherText:   hex.EncodeToString(cipherText),
		CipherParams: cipherParamsJSON,
		KDF:          "pbkdf2",
		KDFParams:    scryptParamsJSON,
		MAC:          hex.EncodeToString(mac),
	}

	store := &EncryptedKey{
		Crypto:  cryptoStruct,
		Id:      id.String(),
		Version: 1,
	}

	jsonString, _ := json.Marshal(store)
	err = ioutil.WriteFile(path, jsonString, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return nil
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}
