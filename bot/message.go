package bot

import "fmt"

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
1. Put /setup_config to configure your server
2. Put /setup_your_bot to deploy your bot to your server.
2. Put /setup_infura to deploy infura services into your server
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
