package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type NodeConfig struct {
	Network          string
	Moniker          string
	BootstrapNode    string
	CoinA            string
	CoinB            string
	RewardAddressBTC string
	RewardAddressETH string
	RewardAddressBNB string
	BlockBookBTC     string
	BlockBookETH     string
	StakeAddr        string
	StakeTx          string
	Memo             string
	KeygenUntil      string
	IsTestnet        bool
}

func NewNodeConfig() *NodeConfig {
	initTime := time.Date(2014, time.December, 31, 12, 13, 24, 0, time.UTC)
	nConf := &NodeConfig{
		CoinA:         "BTC",
		CoinB:         "BTCB",
		BlockBookBTC:  blockBookBTC,
		BlockBookETH:  blockBookETH,
		KeygenUntil:   initTime.Format(time.RFC3339),
		BootstrapNode: "51.158.121.128:12121",
		Network:       networks["1"],
		Moniker:       "Default Node",
	}
	return nConf
}

func (n *NodeConfig) storeConfig(network string, threshold int, members int) error {
	pConfigFileName := fmt.Sprintf("%s/config.toml", network)
	newBaseConfig := strings.ReplaceAll(baseConfig, "**node_moniker_placeholder**", n.Moniker)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_A**", n.CoinA)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_B**", n.CoinB)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**is_testnet**", fmt.Sprintf("%t", n.IsTestnet))

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**threshold_placeholder**", fmt.Sprintf("%d", threshold))
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**participants_placeholder**", fmt.Sprintf("%d", members))

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**btc_blockbook_endpoint**", n.BlockBookBTC)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_address_btc**", n.RewardAddressBTC)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**eth_blockbook_endpoint**", n.BlockBookETH)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_address_eth**", n.RewardAddressETH)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**rpc_uri_placeholder**", bnbSeedNodes[0])
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_tx**", n.StakeTx)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_addr**", n.StakeAddr)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_addr_bnb**", n.RewardAddressBNB)

	newConfigToml := fmt.Sprintf("%s\n", newBaseConfig)
	log.Info(newConfigToml)
	if err := ioutil.WriteFile(pConfigFileName, []byte(newConfigToml), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (n *NodeConfig) saveConfig() error {
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/nConf.json", dataPath), []byte(data), 0600)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeConfig) loadConfig() error {
	str, err := ioutil.ReadFile(fmt.Sprintf("%s/nConf.json", dataPath))
	if err != nil {
		return err
	}
	err = json.Unmarshal(str, &n)
	if err != nil {
		return err
	}
	return nil
}
