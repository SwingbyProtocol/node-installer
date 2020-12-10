package bot

import "fmt"

func makeHelloText() string {
	text := fmt.Sprintf(`
Hello 😊, This is a deploy bot
Steps are here. 
1. Put /setup_server_config to configure your server
2. Put /setup_your_bot to deploy your bot to your server.
4. Put /setup_node to configure your node
3. Put /deploy_infura to deploy infura services into your server
5. Put /deploy_node to deploy your node
	`)
	return text
}

func makeHostText() string {
	text := fmt.Sprintf(`
OK. 
Please let me know your server IP address (Only accept Version 4)
[Configuration step 1/2]
	`)
	return text
}

func (b *Bot) setupIPAndAskUsernameText() string {
	text := fmt.Sprintf(`
OK. Your server IP is %s, 
[Configuration step 2/2]
Please put your username to login into your server.

now: <b>%s</b>

if you want to skip, type 'none'
`, b.nodeIP, b.hostUser)
	return text
}

func (b *Bot) setupUsernameAndLoadSSHkeyText() string {
	text := fmt.Sprintf(`
OK. Your server username is <b>%s</b>
...
SSH_KEY is loaded. Your server is ready. 
Let's setup your bot => /setup_your_bot
`, b.hostUser)
	return text
}

func doneSSHKeyText() string {
	text := fmt.Sprintf(`
OK. SSH_KEY is loaded. Your server is ready. 
Let's setup your bot => /setup_your_bot
`)

	return text
}

func makeDeployBotMessage() string {
	text := fmt.Sprintf(`
OK. Starting deployment... BOT is moving out to your server....

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
BOT is moved out to your server! please go ahead with /setup_node
	`)
	return text
}

func (b *Bot) makeNodeText() string {
	text := fmt.Sprintf(`
OK. 
Next step is you can generate node config
Please put network number from following list.

now: <b>%s</b>

1) BTC --- Binance chain (mainnet)
2) BTC --- Ethereum (mainnet)

3) BTC --- Binance chain (testnet) 
4) BTC --- Ethereum (goerli)

[Configuration step 1/6]
`, b.network)
	return text
}

func (b *Bot) makeUpdateMoniker() string {
	text := fmt.Sprintf(`
OK. What is your Node moniker?

now: <b>%s</b>

[Configuration step 2/6]
if you want to skip, type 'none'
default will be set 'Default Node'
`, b.moniker)
	return text
}

func (b *Bot) makeRewardAddressBTC() string {
	text := fmt.Sprintf(`
OK. Please put your BTC reward address. 
now: <b>%s</b>
[Configuration step 3/6]
if you want to skip, type 'none'
`, b.rewardAddressBTC)

	return text
}

func (b *Bot) makeRewardAddressBNB() string {
	text := fmt.Sprintf(`
OK. Please put your BNB reward address. 
now: <b>%s</b>
[Configuration step 4/6]
if you want to skip, type 'none'
`, b.rewardAddressBNB)
	return text
}

func (b *Bot) makeRewardAddressETH() string {
	text := fmt.Sprintf(`
OK. Please put your ETH reward address. 
now: <b>%s</b>
[Configuration step 5/6]
if you want to skip, type 'none'
`, b.rewardAddressETH)

	return text
}

func makeStakeTxText(addr string, memo string) string {
	text := fmt.Sprintf(`
OK. Your new wallet is generated.

Your address: <b>%s</b>

You have to make stake tx to above address. Please make a tx 

with memo:

<b>%s</b>

Send a timelock transaction to yourself with at least 1,000,000 SWINGBY 

and take note of the transaction ID. Use our portal: https://timelock.swingby.network
[Configuration step 6/6]
`, addr, memo)
	return text
}

func (b *Bot) askStakeTxText() string {
	text := fmt.Sprintf(`
Your staking tx is:
now: <b>%s</b>
Could you put your stake tx hash?
if you want to skip, type 'none'
	`, b.stakeTx)
	return text
}

func (b *Bot) makeStoreKeyText() string {
	text := fmt.Sprintf(`
OK. Setup your new wallet... 
`)
	return text
}

func doneConfigGenerateText() string {
	text := fmt.Sprintf(`
Congratulations!
Your Node configs are updated. 
Let's start deploy => /deploy_node
	`)
	return text
}

func makeDeployNodeMessage() string {
	text := fmt.Sprintf(`
Node upgrading....
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
