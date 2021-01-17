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

	// pEntropy, err := bip39.NewEntropy(256)
	// if err != nil {
	// 	return false, err
	// }
	// // Gen a new address from new entropy
	// pMnemonic, err := bip39.NewMnemonic(pEntropy)
	// if err != nil {
	// 	return false, err
	// }
	// if b.nConf.IsTestnet {
	// 	types.Network = types.TestNetwork
	// }
	// pKey, err := keys.NewMnemonicKeyManager(pMnemonic)
	// if err != nil {
	// 	return false, err
	// }
	// log.Info(pMnemonic)
	// b.nConf.StakeAddr = pKey.GetAddr().String()
	// password, _ := generateRandomBytes(24)
	// log.Infof("Deposit address: %s, pass: %s", b.nConf.StakeAddr, hex.EncodeToString(password))
	// keydata, err := pKey.ExportAsKeyStore(hex.EncodeToString(password))
	// if err != nil {
	// 	return false, err
	// }
	// data, _ := json.Marshal(keydata)
	// err = ioutil.WriteFile(stakeKeyPath, data, 0660)
	// if err != nil {
	// 	return false, err
	// }
	// // Check keystore
	// _, err = keys.NewKeyStoreKeyManager(stakeKeyPath, hex.EncodeToString(password))
	// if err != nil {
	// 	return false, err
	// }
	// b.nConf.Memo = fmt.Sprintf("%s,%s", pP2PKeyHex, b.nConf.RewardAddressETH)
	// return false, nil
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
