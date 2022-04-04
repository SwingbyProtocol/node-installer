package keystore

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/perlin-network/noise/edwards25519"
	"github.com/perlin-network/noise/skademlia"
	log "github.com/sirupsen/logrus"
)

const (
	FileNameLegacy = "keystore.json"
	FileNameBackup = "keystore0_bak.bin"
	FileName       = "keystore0.bin"
	FilePerm       = 0600

	// Skademlia key params
	DefaultSKademliaC1 = 1
	DefaultSKademliaC2 = 1
)

type (
	SaveData struct {
		P2pData        P2PSaveData                `json:"p2p_data"`
		EcdsaTssData   *keygen.LocalPartySaveData `json:"ecdsa_data"`
		EcdsaTssParams *tss.Parameters            `json:"ecdsa_params"`
	}

	P2PSaveData struct {
		SK edwards25519.PrivateKey `json:"private_key"`
		C1 int                     `json:"c1"`
		C2 int                     `json:"c2"`
	}
)

func NewSaveData(p2pKey P2PSaveData, tssParams *tss.Parameters, tssData *keygen.LocalPartySaveData) SaveData {
	return SaveData{P2pData: p2pKey, EcdsaTssParams: tssParams, EcdsaTssData: tssData}
}

func NewP2PSaveData(sk edwards25519.PrivateKey, c1 int, c2 int) P2PSaveData {
	return P2PSaveData{SK: sk, C1: c1, C2: c2}
}

func ReadFromHome(secretHex string, basePath string, optPath ...string) (*SaveData, error) {
	if legacyFileExists(basePath) {
		// convert legacy file to the new encrypted format, rm legacy
		log.Infof("Legacy keystore %s detected; converting to the new format %s and deleting %s.",
			FileNameLegacy, FileName, FileNameLegacy)
		data, err := readFromHomeLegacy(basePath, optPath...)
		if err != nil {
			return nil, err
		}
		if err = WriteToHome(data, secretHex, basePath); err != nil {
			return nil, err
		}
		if err = WriteToHome(data, Path(basePath, FileNameBackup), secretHex); err != nil {
			return nil, err
		}
		oldPath := Path(basePath, FileNameLegacy)
		if err = os.Remove(oldPath); err != nil {
			return nil, err
		}
		return data, nil
	}
	return readFromHomeNew(secretHex, basePath, optPath...)
}

func WriteToHome(data *SaveData, secretHex string, basePath string, optPath ...string) error {
	path := Path(basePath)
	if 0 < len(optPath) {
		path = optPath[0]
	}

	bytes, err := encryptKeyStore(secretHex, data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, FilePerm)
}

func GenerateInHome(secretHex string, basePath string, optPath ...string) error {
	// if there is a legacy keystore file then load the p2p data from it, otherwise generate from scratch
	var p2pData P2PSaveData
	if legacyFileExists(basePath) {
		// take p2p keys from legacy keystore file
		legacyData, err2 := readFromHomeLegacy(Path(basePath, FileNameLegacy))
		if err2 != nil {
			return err2
		}
		p2pData = legacyData.P2pData
	} else {
		keys, c1, c2, err2 := generateP2PKeys()
		if err2 != nil {
			return err2
		}
		p2pData = NewP2PSaveData(keys.PrivateKey(), c1, c2)
	}
	saveData := NewSaveData(p2pData, nil, nil) // tss data populated on keygen
	return WriteToHome(&saveData, secretHex, basePath, optPath...)
}

func LoadOrGenerate(secretHex string, basePath string) (data *SaveData, generated bool, err error) {
	ks, err := ReadFromHome(secretHex, basePath)
	if err != nil {
		if err = GenerateInHome(secretHex, basePath); err != nil {
			return nil, false, err
		}
		ks, err = ReadFromHome(secretHex, basePath) // ensure that it can be read
		if err != nil {
			return nil, false, err
		}
		generated = true
	}
	return ks, generated, err
}

func Path(path string, optFileName ...string) string {
	fileName := FileName
	if 0 < len(optFileName) {
		fileName = optFileName[0]
	}
	return filepath.Join(path, fileName)
}

// ----- //

func readFromHomeNew(secretHex string, basePath string, optPath ...string) (*SaveData, error) {
	path := Path(basePath)
	if 0 < len(optPath) {
		path = optPath[0]
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return decryptKeyStore(secretHex, bytes)
}

func readFromHomeLegacy(basePath string, optPath ...string) (*SaveData, error) {
	path := Path(basePath, FileNameLegacy)
	if 0 < len(optPath) {
		path = optPath[0]
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data := SaveData{}
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func legacyFileExists(basePath string) bool {
	if _, err := os.Stat(Path(basePath, FileNameLegacy)); os.IsNotExist(err) {
		return false
	}
	return true
}

func generateP2PKeys() (keys *skademlia.Keypair, c1, c2 int, err error) {
	c1, c2 = DefaultSKademliaC1, DefaultSKademliaC2
	keys, err = skademlia.NewKeys(c1, c2)
	return
}
