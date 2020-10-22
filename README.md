# node-installer

## Requirements

- Server OS Ubuntu 14.04 LTS
- Disk space ~100GB for chaosnet
 
## Getting Started
- Install `Docker` from https://docker.io to your machine.
- Talk to [@BotFather](https://t.me/BotFather) to create new your `telegram bot` and get your `BOT_TOKEN`
- Setup your instance of your cloud service provider and get ip address and ssh private key
- Run `$ export BOT_TOKEN=<your bot token>`
- Run `$ chmod +x scripts/start_install.sh && scripts/start_install.sh` for Mac User
- Talk to your `telegram bot` with /start command to start install your node.

## Development 
```
$ export BOT_TOKEN=<your bot token>
$ go run main.go
```

## Build
```
$ make build
```