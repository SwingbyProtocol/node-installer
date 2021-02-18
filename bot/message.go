package bot

import "fmt"

func (b *Bot) makeHelloText() string {
	text := fmt.Sprintf(`
Hello 😊
This is <b>Swingby node-installer bot</b>
You can install your meta node and manage node via this bot.

[Setup Node]
/setup_server_config 
 |- configure your server
/setup_your_bot 
 |- move out your bot to your server.

[Version]
Swingby Node: <b>v%s</b>
This Bot: <b>v%s</b>
`, b.nodeVersion, b.botVersion)
	if b.isRemote {
		text = fmt.Sprintf(`
Hello 😊
This is <b>Swingby node-installer bot</b>
You can install your Swingby node and manage it via this bot.

[Setup Node]
/setup_node 
 |- configure your node
 
[Deploy Node]
/deploy_node 
 |- deploy your node
/setup_domain 
 |- configure domain
/enable_domain 
 |- attach domain to your server
/stop_node
 |- stop your node

[Deploy Infura]
/setup_infura 
 |- setup infura containers
/deploy_infura 
 |- deploy infura services

[System management]
/check_status 
 |- checking status of system
/upgrade_bot 
 |- upgrade bot to latest version 
/get_node_logs
 |- getting the latest logs of node

[Version]
Swingby Node: <b>v%s</b>
Bot: <b>v%s</b>
	`, b.nodeVersion, b.botVersion)
	}
	return text
}

func (b *Bot) makeSetupIPText() string {
	text := fmt.Sprintf(`
OK. 
Could you reply your server IP address 
(Only support IPv4)
[Configuration step 1/2]
	`)
	return text
}

func (b *Bot) setupIPAndAskUserNameText() string {
	text := fmt.Sprintf(`
OK. Your server IP is %s, 
[Configuration step 2/2]
Please put your Server Login Username 

now: <b>%s</b>

if you want to skip, type 'none'
`, b.nodeIP, b.hostUser)
	return text
}

func (b *Bot) setupUsernameAndLoadSSHkeyText() string {
	text := fmt.Sprintf(`
OK. Your server Login Username is 

<b>%s</b>

...SSH_KEY file is loaded. Your server is ready. 
Let's setup your bot => /setup_your_bot
`, b.hostUser)
	return text
}

func makeDeployBotMessage() string {
	text := fmt.Sprintf(`
OK. Starting deployment... 
Your bot is moving out to your server....
	`)
	return text
}

func rejectDeployBotByDiskSpaceMessage() string {
	text := fmt.Sprintf(`
Oh sorry. 
The server hasn't enough Disk space on "/var/swingby" mount path.
Minimum <b>1.5TB</b> space required to install Swingby node.
	`)
	return text
}

func errorDeployBotMessage() string {
	text := fmt.Sprintf(`
Oh something wrong. Please kindly check your
IP address, login Username and SSH private key.
	`)
	return text
}

func doneDeployBotMessage() string {
	text := fmt.Sprintf(`
Your bot is moved out to your server! 
Please go ahead with /setup_node
	`)
	return text
}

func makeUpgradeBotMessage() string {
	text := fmt.Sprintf(`
OK. Upgrading your bot....
	`)
	return text
}

func (b *Bot) doneUpgradeBotMessage() string {
	text := fmt.Sprintf(`
System has been upgraded! 
You can start with /start command.
`)
	return text
}

func (b *Bot) makeNodeText() string {

	// 2) BTC --- Binance chain (mainnet)

	// 3) BTC --- Binance chain (testnet)
	// 4) BTC --- Ethereum (goerli)
	text := fmt.Sprintf(`
OK. 
This steps generates node config
Please put target network number on following list.

now: <b>%s</b>

1) BTC --- Ethereum (mainnet)

[Configuration step 1/4]
`, b.nConf.Network)
	return text
}

func (b *Bot) makeUpdateMoniker() string {
	text := fmt.Sprintf(`
OK. What is your Node moniker?

now: <b>%s</b>

[Configuration step 2/4]
if you want to skip, type 'none'
default will be set 'Default Node'
`, b.nConf.Moniker)
	return text
}

func (b *Bot) makeRewardAddressBNB() string {
	text := fmt.Sprintf(`
OK. Please put your BNB reward address. 
now: <b>%s</b>
[Configuration step 4/6]
if you want to skip, type 'none'
`, b.nConf.RewardAddressBNB)
	return text
}

func (b *Bot) makeRewardAddressETH() string {
	text := fmt.Sprintf(`
OK. Please put your ETH reward address. 
now: <b>%s</b>
[Configuration step 3/4]
if you want to skip, type 'none'
`, b.nConf.RewardAddressETH)

	return text
}

func (b *Bot) makeStakeAddrText() string {
	text := fmt.Sprintf(`
OK. Your new p2p node key is generated.

You have to make a stake tx. 
Following steps:
1. Setup your BNB wallet: https://www.binance.org/en/create
2. Access our timelock portal: https://timelock.swingby.network
3. Make a "timelock" tx with this "description"
4. Put your "staking" BNB wallet address.

description:

<b>%s</b>

<b>Note: minimum stake amount is least 150,000 SWINGBYs with 1 month lock</b>
`, b.nConf.Memo)
	return text
}

func (b *Bot) askStakeAddrText() string {
	text := fmt.Sprintf(`
Your staking BNB address is:

now: <b>%s</b>

Could you put your BNB staking address?
[Configuration step 4/4]
if you want to skip, type 'none'
	`, b.nConf.StakeAddr)
	return text
}

func (b *Bot) makeStoreKeyText() string {
	text := fmt.Sprintf(`
OK. Setup your p2p node keys... 
`)
	return text
}

func doneConfigGenerateText() string {
	text := fmt.Sprintf(`
Congratulations!
Your Swingby node config has been updated. 

Next step is installing infura package.
Let's start => /setup_infura.
	`)
	return text
}

func (b *Bot) setupDomainText() string {
	text := fmt.Sprintf(`
OK. 
Please put your subdomain like 

testnode-1.example.com

now subdomain is:

<b>%s</b>

if you want to skip, type 'none'
`, b.nConf.Domain)
	return text
}

func (b *Bot) doneDomainText() string {
	text := fmt.Sprintf(`
OK. Your server subdomain is 

<b>%s</b>

IP address is

<b>%s</b>

You have to attach domain A record to your server before use
/enable_domain
`, b.nConf.Domain, b.nodeIP)
	return text
}

func (b *Bot) makeDomainMessage() string {
	text := fmt.Sprintf(`
Your subdomain will be attached to your server

<b>%s</b> 

to

<b>%s</b>

Deploying Nginx....
`, b.nConf.Domain, b.nodeIP)
	return text
}

func (b *Bot) doneDomainMessage() string {
	text := fmt.Sprintf(`
Your subdomain has been attached. 
Let's access https://%s

	`, b.nConf.Domain)
	return text
}

func errorDomainMessage() string {
	text := fmt.Sprintf(`
You subdomain is not attahced. Please kindly check error logs
	`)
	return text
}
func (b *Bot) makeDeployNodeMessage() string {
	text := fmt.Sprintf(`
Deploying your Swingby node.... (v%s)
	`, b.nodeVersion)
	return text
}

func rejectDeployNodeByInfuraMessage() string {
	text := fmt.Sprintf(`
This command is not avaialbe now.
Infura syncing should be 100.00%% done
Please try /check_status first.
`)
	return text
}

func rejectDeployNodeByConfigMessage() string {
	text := fmt.Sprintf(`
This command is not avaialbe now.
You have to node config.
Please try /setup_node first.
`)
	return text
}

func doneDeployNodeMessage() string {
	text := fmt.Sprintf(`
Your Swingby node has been deployed! 
(Updated to latest version)
	`)
	return text
}

func errorDeployNodeMessage() string {
	text := fmt.Sprintf(`
Deployment is not completed. Please kindly check error logs
	`)
	return text
}

func (b *Bot) makeStopNodeMessage() string {
	text := fmt.Sprintf(`
Stopping your Swingby node.... (v%s)
	`, b.nodeVersion)
	return text
}

func (b *Bot) errorStopNodeMessage() string {
	text := fmt.Sprintf(`
Something wrong. Please kindly check error logs
	`)
	return text
}

func (b *Bot) doneStopNodeMessage() string {
	text := fmt.Sprintf(`
Your Swingby node has been stopped!
`)
	return text
}

func confirmSetupInfuraMessage() string {
	text := fmt.Sprintf(`
<b>This command removes your blockchain data.</b>
And blockchain data will be rollback to latest snapshot.
If you sure this, please go ahead /setup_infura
`)
	return text
}

func makeSetupInfuraMessage() string {
	text := fmt.Sprintf(`
Installing infura packages...
`)
	return text
}

func doneSetupInfuraMessage() string {
	text := fmt.Sprintf(`
Syncing of the snapshot data....
(This process may takes long time...)
You can check the syncing progress by /check_status
	`)
	return text
}

func errorSetupInfuraMessage() string {
	text := fmt.Sprintf(`
Someting wrong. Please kindly check error logs
	`)
	return text
}

func rejectDeployInfuraMessage() string {
	text := fmt.Sprintf(`
Syncing progress is not completed yet.
You can check the syncing progress by /check_status first.
`)
	return text
}

func confirmDeployInfuraMessage() string {
	text := fmt.Sprintf(`
This command will restarts geth nodes.
<b>it may takes a long time to sync blockchain again. </b>
if you sure this, please go ahead /deploy_infura
`)
	return text
}

func makeDeployInfuraMessage() string {
	text := fmt.Sprintf(`
Deploying infura containers....
`)
	return text
}

func doneDeployInfuraMessage() string {
	text := fmt.Sprintf(`
All infura containers are upgraded!
Status check => /check_status
	`)
	return text
}

func errorDeployInfuraMessage() string {
	text := fmt.Sprintf(`
Deployment has been rejected. Please kindly check error logs.
	`)
	return text
}

func makeCheckNodeMessage() string {
	text := fmt.Sprintf(`
Getting the latest node status...
`)
	return text
}

func (b *Bot) checkNodeMessage() string {
	b.mu.RLock()
	text := fmt.Sprintf(`
[Syncing status]
<b>%.2f%%</b> finished.

[Blockchain syncing status]
BTC: <b>#%d</b> (%.3f%%)
ETH: <b>#%d</b> (%.3f%%)

After reached 99.99%% of progress,
You can start deploy infura containers by
/deploy_infura

After reached 100.00%% of progress,
You can install node by 
/deploy_node
`, b.syncProgress, b.bestHeight["BTC"], b.SyncRatio["BTC"], b.bestHeight["ETH"], b.SyncRatio["ETH"])
	b.mu.RUnlock()
	return text
}

func errorCheckNodeMessage() string {
	text := fmt.Sprintf(`
Node data checking is failed, could you try it later.
	`)
	return text
}

func errorLogFileMessage() string {
	text := fmt.Sprintf(`
Error: Log file is not exist...
`)
	return text
}

func upgradeBotMessage(newVersion string) string {
	text := fmt.Sprintf(`
The new bot [v%s] is released!
Let's upgrade /upgrade_bot

And then, Let's deploy node again.
/deploy_node
	`, newVersion)
	return text
}

func upgradeNodeMessage(newVersion string) string {
	text := fmt.Sprintf(`
The new node [v%s] is released!
Let's deploy again /deploy_node
	`, newVersion)
	return text
}

func upgradeBothMessage(newBotVersion string, newNodeVersion string) string {
	text := fmt.Sprintf(`
The bot[v%s] and node [v%s] is released!

Let's upgrade /upgrade_bot

then, Let's deploy again by
/check_status and /deploy_node
	`, newBotVersion, newNodeVersion)
	return text
}
