package bot

import "fmt"

func makeHelloText() string {
	text := fmt.Sprintf(`
Hello 😊, This is a deploy bot
Steps is here. 
1. Put /setup_config to configure your server
2. Put /setup_your_bot to deploy your bot to your server.
3. Put /deploy_infura to deploy infura services into your server
4. Put /setup_node to configure your node
5. Put /deploy_node to deploy your node
	`)
	return text
}

func makeHostText() string {
	text := fmt.Sprintf(`
Cool. 
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
Cool. SSH_KEY is loaded. Your server is ready. 
Let's setup your bot => /setup_your_bot
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
BOT is moved out to your server! please go ahead with /setup_node
	`)
	return text
}

func (b *Bot) makeNodeText() string {
	text := fmt.Sprintf(`
Cool. 
Next step is you can generate node config
Please put network number from following list.

now: <b>%s</b>

1) BTC --- Binance chain testnet 
2) BTC --- Ethereum testnet (goerli)
3) BTC --- Binance Smart Chain testnet
[Configuration step 1/6]
`, b.network)
	return text
}

func (b *Bot) makeUpdateMoniker() string {
	text := fmt.Sprintf(`
Cool. What is your Node moniker?

now: <b>%s</b>

[Configuration step 2/6]
if you want to skip, type 'none'
default will be set 'Default Node'
`, b.moniker)
	return text
}

func (b *Bot) makeRewardAddressBTC() string {
	text := fmt.Sprintf(`
Cool. Please put your BTC reward address. 
now: <b>%s</b>
[Configuration step 3/6]
if you want to skip, type 'none'
`, b.rewardAddressBTC)

	return text
}

func (b *Bot) makeRewardAddressBNB() string {
	text := fmt.Sprintf(`
Cool. Please put your BNB reward address. 
now: <b>%s</b>
[Configuration step 4/6]
if you want to skip, type 'none'
`, b.rewardAddressBNB)
	return text
}

func (b *Bot) makeRewardAddressETH() string {
	text := fmt.Sprintf(`
Cool. Please put your ETH reward address. 
now: <b>%s</b>
[Configuration step 5/6]
if you want to skip, type 'none'
`, b.rewardAddressETH)

	return text
}

func makeStakeTxText(addr string, memo string) string {
	text := fmt.Sprintf(`
Cool. Your new wallet is generated. and new mnemonic is stored into server.

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

func askStakeTxText() string {
	text := fmt.Sprintf(`
What is your Transaction ID? Please put your stake tx Hash
	`)
	return text
}

func (b *Bot) makeStoreKeyText() string {
	text := fmt.Sprintf(`
Cool. Generating new your wallet and store your new mnemonic key... 
`)
	return text
}

func doneConfigGenerateText() string {
	text := fmt.Sprintf(`
Congratulations!
Your Node configs are updated.
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
