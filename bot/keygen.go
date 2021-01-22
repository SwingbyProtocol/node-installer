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
	"github.com/binance-chain/go-sdk/keys"
)

func (b *Bot) generateKeys(basePath string) (bool, error) {
	pDataDirName := fmt.Sprintf("%s/data", basePath)
	pKeystoreFileName := fmt.Sprintf("%s/keystore.json", pDataDirName)
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

	b.nConf.Memo = fmt.Sprintf("%s,%s", pP2PKeyHex, b.nConf.RewardAddressETH)
	return true, nil
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
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
