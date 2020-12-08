# node-installer

## Requirements

- OS Ubuntu 20.04 LTS
- vCPUs  4~
- Memory 8GB~
- Disk space ~100GB for Chaosnet
 
## Getting Started
- Install `Docker` from https://docker.io to your machine.
- Talk to [@BotFather](https://t.me/BotFather) to create new your `telegram bot` and get your `BOT_TOKEN`
- Setup your instance of your cloud service provider and get ip address and ssh private key
- Store your ssh private key to accesss instance into `data/ssh_key` 
- Run `$ export BOT_TOKEN=<your bot token>`
- Run `$ chmod +x scripts/start_build_and_install.sh && scripts/start_build_and_install.sh` for Mac User
- Talk to your `telegram bot` with `/start` command to start install your node.

## Development 
```
$ export BOT_TOKEN=<your bot token>
$ go run main.go
```

## Build
```
$ make build
```
