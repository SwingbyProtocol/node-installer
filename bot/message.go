package bot

import "fmt"

func (b *Bot) makeHelloText() string {
	text := fmt.Sprintf(`
Hello 😊
This is <b>Swingby node-installer bot</b>
You can setup your node through this bot.

[Setup]
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
You can setup your node through this bot.

[Setup]
/setup_node 
 |- configure your node
/show_timelock_memo
 |- view your timelock memo
/show_p2pkey
 |- view your p2p key for ERC20 staking

[Infura management]
/setup_infura [not recommended]
 |- setup infura containers
/resync_infura [not recommended]
 |- re-syncing snapshot
/remove_infura
 |- remove infura data
/reset_geth
 |- remove data (only geth) & restart geth
/deploy_infura 
 |- deploy infura services
/set_global_infura
 |- use foundation infura
/set_local_infura
 |- use local infura

[Node management]
/deploy_node 
 |- deploy your node
/stop_node
 |- stop your node
/get_node_logs
 |- getting the latest logs of node
 
[Domain management]
/setup_domain 
 |- configure domain
/enable_domain 
 |- attach domain to your server
/stop_nginx
 |- stop nginx

[Server management]
/check_status 
 |- checking status of system
/upgrade_bot 
 |- upgrade bot to latest version 

[Version]
Swingby Node: <b>v%s</b>
Bot: <b>v%s</b>
	`, b.nodeVersion, b.botVersion)
	}
	return text
}

func (b *Bot) instructionMention() string {
	text := fmt.Sprint(`
Pelase confirm these steps before starting this bot

1) check whether you choosed OS <b>Ubuntu 20.04-LTS</b> for your server.

2) check whether the <b>/var/swingby</b> is exist on your server because we will deploy this bot into there.

3) check whether the ssh key is valid <b>(./data/ssh_key)</b> because bot may can't access into your server correctlly.
you can try this if this problem is exist.

$ chmod 0600 ./data/ssh_key

`)
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

func rejectDeployInfuraByDiskSpaceMessage() string {
	text := fmt.Sprintf(`
Oh sorry. 
The server hasn't enough Disk space on "/var/swingby" mount path.

- Minimum <b>1.5TB for BTC-ETH network </b> 
- Minimum <b>1.6TB for BTC-BSC network </b> 

space required to install Swingby node.
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
3) BTC --- ETH (skypool)

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
OK. Please put your ETH/BSC reward address. 
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

Following steps for Skypool setup/migration
1. Swap BEP20 SWINGBY to ERC20 in https://bridge.swingby.network
2. do /show_p2pkey
3. copy that NodeP2PKey
4. put NodeP2PKey in "Node Stake" view on https://dao.swingby.network 
Tutorial is here
https://skybridge-docs.swingby.network/swingby-dao/tutorials/bond-metanodes

Note: stake amount is least 175,000 SWINGBYs 
with over 1 month timelock
(recommended: at least 3 months)

`)
	return text
}

func (b *Bot) askStakeAddrText() string {
	text := fmt.Sprintf(`
Your staking BNB/ETH address is:

now: <b>%s</b>

Could you put your BNB/ETH staking address?

<b>[important]</b>
for skypool you have to put <b>ETH</b> staking address.

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

func (b *Bot) showMemoText(memo string, stakeAddr string) string {
	text := fmt.Sprintf(`
timelock_memo:

<b>%s</b>

address: <b>%s</b>
`, memo, stakeAddr)
	return text
}

func (b *Bot) showP2PKeyText(memo string, stakeAddr string) string {
	text := fmt.Sprintf(`
NodeP2PKey:

<b>%s</b>

your ETH addr: <b>%s</b>
`, memo[0:64], stakeAddr)
	return text
}

func doneConfigGenerateText() string {
	text := fmt.Sprintf(`
Congratulations!
Your Swingby node config has been updated. 

Next step is installing infura package.
Let's start => /deploy_infura.
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
	nodeV := b.nextNodeVersion
	if b.nConf.Network == Network3 {
		nodeV = "0.1.3-sp"
	}
	text := fmt.Sprintf(`
Deploying your Swingby node.... (v%s)
	`, nodeV)
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

func rejectDeployNodeByUpgradeInfuraMessage() string {
	text := fmt.Sprintf(`
This command is not avaialbe now.
need to upgrade new geth
Please try /deploy_infura first.
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

func (b *Bot) makeStopNginxMessage() string {
	text := fmt.Sprintf(`
Stopping your Nginx....
	`)
	return text
}

func (b *Bot) errorStopNginxMessage() string {
	text := fmt.Sprintf(`
Something wrong. Please kindly check error logs
	`)
	return text
}

func (b *Bot) doneStopNginxMessage() string {
	text := fmt.Sprintf(`
Your Nginx has been stopped!
`)
	return text
}

func (b *Bot) makeUpdateStakeNodeMessage() string {
	text := fmt.Sprintf(`
Upgrade your Swingby node.... (v%s)
	`, b.nodeVersion)
	return text
}

func (b *Bot) errorUpdateStakeNodeMessage() string {
	text := fmt.Sprintf(`
Upgrade wrong. Please kindly check error logs
	`)
	return text
}

func (b *Bot) doneUpdateStakeNodeMessage() string {
	text := fmt.Sprintf(`
Your Swingby node has been upgrade!
`)
	return text
}

func makeResyncInfuraMessage() string {
	text := fmt.Sprintf(`
Re-syncing infura packages...
`)
	return text
}

func confirmResyncInfuraMessage() string {
	text := fmt.Sprintf(`
<b>This command removes your blockchain data.</b>
Blockchain data will be rollback to latest snapshot.
If you are sure about this, please go ahead /resync_infura
`)
	return text
}

func doneResyncInfuraMessage() string {
	text := fmt.Sprintf(`
Re-Syncing of the snapshot data....
(This process may take a long time...)
You can check the syncing progress by /check_status
	`)
	return text
}

func errorResyncInfuraMessage() string {
	text := fmt.Sprintf(`
Someting is wrong. Please kindly check error logs
	`)
	return text
}

func makeRemoveInfuraMessage() string {
	text := fmt.Sprintf(`
Removing infura packages...
`)
	return text
}

func confirmRemoveInfuraMessage() string {
	text := fmt.Sprintf(`
<b>This command removes your blockchain data.</b>
Blockchain data will be rollback to latest snapshot.
If you are sure about this, please go ahead /remove_infura
`)
	return text
}

func doneRemoveInfuraMessage() string {
	text := fmt.Sprintf(`
Removing dir has been completed. (/var/swingby/mainnet) 
	`)
	return text
}

func errorRemoveInfuraMessage() string {
	text := fmt.Sprintf(`
Someting is wrong. Please kindly check error logs
	`)
	return text
}

func makeSetupInfuraMessage() string {
	text := fmt.Sprintf(`
Installing infura packages...
`)
	return text
}

func confirmSetupInfuraMessage() string {
	text := fmt.Sprintf(`
<b>This command removes your blockchain data.</b>
Blockchain data will be rolled back to the latest snapshot.
If you are sure about this, please go ahead /setup_infura
`)
	return text
}

func doneSetupInfuraMessage() string {
	text := fmt.Sprintf(`
Syncing of the snapshot data....
(This process may take a long time...)
You can check the syncing progress by /check_status
	`)
	return text
}

func errorSetupInfuraMessage() string {
	text := fmt.Sprintf(`
Someting is wrong. Please check error logs
	`)
	return text
}

func rejectDeployInfuraMessage() string {
	text := fmt.Sprintf(`
Syncing is not completed yet.
You can check the syncing progress by /check_status first.
`)
	return text
}

func confirmDeployInfuraMessage() string {
	text := fmt.Sprintf(`
This command will restart geth nodes.
<b>It may take a long time to sync the blockchain data again.</b>
If you are sure about this, please go ahead.
=> /deploy_infura
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
Deployment has been rejected. Please check error logs.
	`)
	return text
}

func makeResetGethMessage() string {
	text := fmt.Sprintf(`
Removing ETH/BSC data & restarting...
`)
	return text
}

func confirmResetGethMessage() string {
	text := fmt.Sprintf(`
<b>This command removes your ETH/BSC blockchain data.</b>
Blockchain data will be rollback to latest snapshot.
If you are sure about this, please go ahead /reset_geth
`)
	return text
}

func doneResetGethMessage() string {
	text := fmt.Sprintf(`
ETH/BSC data has been removed and ETH/BSC node is restarted.
(Please wait a few hours to sync up blockchains) 
Check status again => /check_status
	`)
	return text
}

func errorResetGethMessage() string {
	text := fmt.Sprintf(`
Someting is wrong. Please kindly check error logs
	`)
	return text
}

func (b *Bot) makeCheckNodeMessage() string {
	text := fmt.Sprintf(`
Getting the latest node status...
NodeIP--<b>%s</b>
`, b.nodeIP)
	return text
}

func (b *Bot) checkNodeMessage(varAvailableBytes int) string {
	b.mu.RLock()
	coinBSymbol := "ETH"
	nodeVersion := GethLockVersion
	switch b.nConf.Network {
	case Network1:
	case Network2:
		coinBSymbol = "BSC"
		nodeVersion = BSCLockVersion
	case Network3:
	}
	availableGBs := varAvailableBytes / 1024

	text := fmt.Sprintf(`
[Syncing status]
<b>%.2f%%</b> finished.

[Blockchain syncing status]
[mode: <b>%s</b>]
BTC: <b>#%d</b> (%.3f%%)
%s: <b>#%d</b> (%.3f%%)
|- target: <b>[#%d]</b>

[Domain status]
Nginx active: <b>%s</b>

[Storage status]
Available space for /var/swingby:
<b>~%d GB</b>

After reached 99.99%% of progress,
You can start deploy infura
/deploy_infura 
[%s: %t]

After reached 100.00%% of progress,
You can install node by 
/deploy_node
`, b.syncProgress, b.infura, b.bestHeight["BTC"], b.SyncRatio["BTC"], coinBSymbol, b.bestHeight["ETH"], b.SyncRatio["ETH"], b.etherScanHeight, b.isActiveNginx, availableGBs, nodeVersion, b.validInfura)
	b.mu.RUnlock()
	return text
}

func informStorageIssue(gb int) string {
	text := fmt.Sprintf(`
[Warning] Hi Human!
Your free disk space is running out. (only %dGB reft)
You can check status => /check_status
You may want to resync to the latest snapshot. (Takes about 4 hours)

To do re-syncing process

1. /reset_geth
2. /check_status

if you'd like to use global infura to keep your node online during this process, you can use these steps:

1. /set_global_infura
2. /deploy_node

if you have any questions, 
Please ask communities in #metanodes room in discord!
`, gb)
	return text
}

func errorCheckNodeMessage() string {
	text := fmt.Sprintf(`
Unable to check the node status, please check the logs.
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

func restartBlockbookMessage() string {
	text := fmt.Sprintf(`
[WARN] 
Hi. Your blockbooks are restarted because syncing is stopped.

Your node may not ONLINE at the moment.
Please watch node status /check_status
	`)
	return text
}

func (b *Bot) BotDownMessage() string {
	text := fmt.Sprintf(`
Hey Human! 
This Bot is down!`)
	return text
}
