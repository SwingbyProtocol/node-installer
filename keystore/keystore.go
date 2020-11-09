package keystore

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/perlin-network/noise/edwards25519"
	"github.com/perlin-network/noise/skademlia"
)

const (
	FileName = "keystore.json"
	FilePerm = 0660

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

func WriteToHome(data SaveData, optPath ...string) error {
	path := ""
	if 0 < len(optPath) {
		path = optPath[0]
	}
	file, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, file, FilePerm)
}

func ReadFromHome(optPath ...string) (*SaveData, error) {
	path := ""
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

func GenerateInHome(optPath ...string) error {
	keys, c1, c2, err := generateP2PKeys()
	if err != nil {
		return err
	}
	p2pData := NewP2PSaveData(keys.PrivateKey(), c1, c2)
	tssPreParams, err := keygen.GeneratePreParams(1 * time.Minute)
	if err != nil {
		return err
	}
	tssSaveData := &keygen.LocalPartySaveData{
		LocalPreParams: *tssPreParams,
	}
	saveData := NewSaveData(p2pData, nil, tssSaveData)
	return WriteToHome(saveData, optPath...)
}

func LoadOrGenerate() (data *SaveData, generated bool, err error) {
	kstore, err := ReadFromHome()
	if err != nil {
		if err = GenerateInHome(); err != nil {
			return nil, false, err
		}
		kstore, err = ReadFromHome() // ensure that it can be read
		if err != nil {
			return nil, false, err
		}
		generated = true
	}
	return kstore, generated, err
}

func generateP2PKeys() (keys *skademlia.Keypair, c1, c2 int, err error) {
	c1 = DefaultSKademliaC1
	c2 = DefaultSKademliaC2
	keys, err = skademlia.NewKeys(c1, c2)
	return
}
