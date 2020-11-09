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
BOT is moving to your server.
	`)
	return text
}

func doneDeployBotMessage() string {
	text := fmt.Sprintf(`
Starting deployment... BOT is moved out to your server....
	`)
	return text
}

func makeHelloText() string {
	text := fmt.Sprintf(`
Hello 😊, This is a deploy bot
Steps is here. 
1. Put /setup_server to configure your server
2. Put /setup_bot to deploy your bot to your server.
2. Put /setup_infura to deploy infura services into your server
	`)
	return text
}

func makeDeployText() string {
	text := fmt.Sprintf(`
Deploy is starting
Please let me know your server IP address (Only accept Version 4)
	`)
	return text
}
