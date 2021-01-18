package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	WalletContract      = "0x0fc2c6513ffc15d92a7593cede8b44cec3d85122"
	WalletContractTest  = "0xf50b87c16bfb0781a86d4a7e91eb9e1da16906c4"
	LPtokenContract     = "0xb7f7dd6D0e3addBb98F8Bc84F010B580A6151b08"
	LPtokenContractTest = "0xf50b87c16bfb0781a86d4a7e91eb9e1da16906c4"
	WBTCContract        = "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"
	WBTCContractTest    = "0xf50b87c16bfb0781a86d4a7e91eb9e1da16906c4"
	BootstrapNode       = "116.203.5.120:12122"
	BootstrapNodeTest   = "51.158.121.128:12121"
	GethRPC             = "10.2.0.1:8545"
	BlockBookBTC        = "10.2.0.1:9130"
	BlockBookETH        = "10.2.0.1:9131"
	StopTrigger         = "https://btc-wbtc-mainnet.s3.eu-central-1.amazonaws.com/platform_status.json"
)

var BnbSeedNodes = []string{
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
preferred_uri = "**node_preferred_uri**"

[logger]
level = "INFO"
max_file_size_MB = 100
max_backup_files = 100
max_retain_files_days = 14
use_console_logger = true
use_file_logger = true
compress = true

[swaps]
testnet = **is_testnet**
coin_1 = "**coin_A**"
coin_2 = "**coin_B**"
stake_coin = "SWINGBY-888"
stop_trigger_uri = "**stop_trigger_uri**"
# (using defaults in code)
# fee_percent = 0.2
# max_amount = 1
# min_amount_swap = 0.0004
# min_amount_refund = 0.001

[tss]
threshold = **threshold_placeholder**
keygen_until = "2020-07-23T12:00:00Z"

[btc]
rest_uri = "http://**btc_blockbook_endpoint**"
ws_uri = "ws://**btc_blockbook_endpoint**/websocket"
# miner_fee = 0.0003

[eth]
rpc_uri = "http://**eth_rpc_endpoint**"
rest_uri = "http://**eth_blockbook_endpoint**"
ws_uri = "ws://**eth_blockbook_endpoint**/websocket"
wallet_contract_addr = "**eth_wallet_contract**"
lp_token_contract_addr = "**eth_lpt_contract**"
btc_token_contract_addr = "**btc_token_contract_addr**"
# miner_fee = 0.00015

[bnb]
rpc_uri = "**rpc_uri_placeholder**"
http_uri = "https://explorer.binance.org"
# miner_fee = 0.000001
stake_addr = "**stake_addr**"
`

type NodeConfig struct {
	Network          string
	Moniker          string
	BootstrapNode    string
	Domain           string
	PreferredURI     string
	BNBSeed          string
	CoinA            string
	CoinB            string
	RewardAddressETH string
	RewardAddressBNB string
	GethRPC          string
	BlockBookBTC     string
	BlockBookETH     string
	StakeAddr        string
	StakeTx          string
	WalletContract   string
	LPtoken          string
	WBTCContract     string
	StopTrigger      string
	Memo             string
	KeygenUntil      string
	IsTestnet        bool
	Threshold        int
	Members          int
}

func NewNodeConfig() *NodeConfig {
	initTime := time.Date(2014, time.December, 31, 12, 13, 24, 0, time.UTC)
	nConf := &NodeConfig{
		CoinA:          "WBTC",
		CoinB:          "BTC",
		GethRPC:        GethRPC,
		BNBSeed:        BnbSeedNodes[0],
		BlockBookBTC:   BlockBookBTC,
		BlockBookETH:   BlockBookETH,
		KeygenUntil:    initTime.Format(time.RFC3339),
		BootstrapNode:  BootstrapNode,
		Network:        Networks["1"],
		Moniker:        "Default Node",
		WalletContract: WalletContract,
		LPtoken:        LPtokenContract,
		WBTCContract:   WBTCContract,
		StopTrigger:    StopTrigger,
	}
	return nConf
}

func (n *NodeConfig) SetMainnet() {
	n.IsTestnet = false
	n.WalletContract = WalletContract
	n.LPtoken = LPtokenContract
	n.BootstrapNode = BootstrapNode
	n.WBTCContract = WBTCContract
}

func (n *NodeConfig) SetTestnet() {
	n.IsTestnet = true
	n.WalletContract = WalletContractTest
	n.LPtoken = LPtokenContractTest
	n.BootstrapNode = BootstrapNodeTest
	n.WBTCContract = WBTCContractTest
}

func (n *NodeConfig) SetDomain(domain string) {
	n.Domain = domain
	n.PreferredURI = fmt.Sprintf("https://%s", domain)
}

func (n *NodeConfig) SetTSSGroup(members int, threshold int) {
	n.Members = members
	n.Threshold = threshold
}

func (n *NodeConfig) storeConfig() error {
	pConfigFileName := fmt.Sprintf("%s/%s/config.toml", dataPath, n.Network)
	newBaseConfig := strings.ReplaceAll(baseConfig, "**node_moniker_placeholder**", n.Moniker)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**node_preferred_uri**", n.PreferredURI)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_A**", n.CoinA)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**coin_B**", n.CoinB)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stop_trigger_uri**", n.StopTrigger)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**is_testnet**", fmt.Sprintf("%t", n.IsTestnet))

	//newBaseConfig = strings.ReplaceAll(newBaseConfig, "**participants_placeholder**", fmt.Sprintf("%d", n.Members))
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**threshold_placeholder**", fmt.Sprintf("%d", n.Threshold))

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**btc_blockbook_endpoint**", n.BlockBookBTC)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**eth_rpc_endpoint**", n.GethRPC)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**eth_blockbook_endpoint**", n.BlockBookETH)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**eth_wallet_contract**", n.WalletContract)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**eth_lpt_contract**", n.LPtoken)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**btc_token_contract_addr**", n.WBTCContract)
	//newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_address_eth**", n.RewardAddressETH)

	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**rpc_uri_placeholder**", n.BNBSeed)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_tx**", n.StakeTx)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**stake_addr**", n.StakeAddr)
	newBaseConfig = strings.ReplaceAll(newBaseConfig, "**reward_addr_bnb**", n.RewardAddressBNB)

	newConfigToml := fmt.Sprintf("%s\n", newBaseConfig)
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
	err = ioutil.WriteFile(fmt.Sprintf("%s/node_config.json", dataPath), []byte(data), 0600)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeConfig) loadConfig() error {
	str, err := ioutil.ReadFile(fmt.Sprintf("%s/node_config.json", dataPath))
	if err != nil {
		return err
	}
	err = json.Unmarshal(str, &n)
	if err != nil {
		return err
	}
	return nil
}
