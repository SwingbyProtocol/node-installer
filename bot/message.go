package bot

import "fmt"

var networks = map[string]string{
	"1": network1,
	"2": network2,
	"3": network3,
}

func makeDeployNodeMessage() string {
	text := fmt.Sprintf(`
Node is upgrading....
	`)
	return text
}

func doneDeployNodeMessage() string {
	text := fmt.Sprintf(`
Node is upgraded.
	`)
	return text
}

func makeDeployInfuraMessage() string {
	text := fmt.Sprintf(`
Infura is upgrading....
	`)
	return text
}

func doneDeployInfuraMessage() string {
	text := fmt.Sprintf(`
Infura is upgraded.
	`)
	return text
}

func makeDeployBotMessage() string {
	text := fmt.Sprintf(`
OK cool. Starting deployment... BOT is moving out to your server....

	`)
	return text
}

func errorDeployBotMessage() string {
	text := fmt.Sprintf(`
Oh something error is happened. Please kindly check your server IP address and SSH key again.
	`)
	return text
}

func doneDeployBotMessage() string {
	text := fmt.Sprintf(`
BOT is moved out to your server! please go ahead with /setup_infura
	`)
	return text
}

func makeHelloText() string {
	text := fmt.Sprintf(`
Hello 😊, This is a deploy bot
Steps is here. 
1. Put /setup_node to configure your node
1. Put /setup_config to configure your server
2. Put /setup_your_bot to deploy your bot to your server.
3. Put /deploy_infura to deploy infura services into your server
4. Put /deploy_node to deploy your node

	`)
	return text
}

func makeHostText() string {
	text := fmt.Sprintf(`
Cool. 
[Configuration step 1/2]
Please let me know your server IP address (Only accept Version 4)
	`)
	return text
}

func seutpSSHKeyText(ip string) string {
	text := fmt.Sprintf(`
Cool. Your server IP is %s, 
[Configuration step 2/2]
Please put your SSH private key.
`, ip)

	return text
}

func doneSSHKeyText() string {
	text := fmt.Sprintf(`
Cool. Your server is ready. 
`)

	return text
}

func makeNodeText() string {
	text := fmt.Sprintf(`
Cool. Your server is ready. 
Next step is you can generate node config
[Configuration step 1/2]
What network will you using? put number.
1) BTC_BTCB testnet 
2) BTC_Ethereum testnet (goerli)
3) BTC_Binance Smart Chain testnet
`)
	return text
}

func makeRewardAddressBTC() string {
	text := fmt.Sprintf(`
Cool. Please put your BTC reward address. 
[Configuration step 2/3]
`)

	return text
}

func makeRewardAddressETH() string {
	text := fmt.Sprintf(`
Cool. Please put your ETH reward address. 
[Configuration step 3/3]
`)

	return text
}

func makeRewardAddressBNB() string {
	text := fmt.Sprintf(`
Cool. Please put your BNB reward address. 
[Configuration step 3/3]
`)

	return text
}

func makeStakeTxText(addr string, memo string) string {
	text := fmt.Sprintf(`
You have to make stake tx. Please make a tx 

to: <b>%s</b>

with memo:

<b>%s</b>

Send a timelock transaction to yourself with at least 1,000,000 SWINGBY 
and take note of the transaction ID. Use our portal: https://timelock.swingby.network
`, addr, memo)
	return text
}

func makeStoreKeyText() string {
	text := fmt.Sprintf(`
Cool. Generating new your wallet and store your new mnemonic key... 
`)
	return text
}
