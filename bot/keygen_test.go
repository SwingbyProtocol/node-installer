package bot

import "testing"

func TestKeygen(t *testing.T) {
	storeConfig("../data", "test_moniker", 12, 24, "BTC", "BTCE", "blockbook-btc", "blockbook-eth:9212", "address_btc", "address_ETH", "address_bnb", "stake_tx")
}
