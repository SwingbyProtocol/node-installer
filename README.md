# node-installer
[![Go Report Card](https://goreportcard.com/badge/github.com/SwingbyProtocol/node-installer)](https://goreportcard.com/report/github.com/SwingbyProtocol/node-installer)

## Requirements (Mainnet environment)

- OS Ubuntu 20.04 LTS
- CPUs >= 4
- Memory >= 16GB
- Disk space >= 1.7TB
- Swap memory >= 5GB
- Network bandwidth >= 500Mbps

The docker host should supports swap memory. [details](https://docs.docker.com/config/containers/resource_constraints/)

To enable swap memory, set up your configuration and restart your instance according to this document.
https://docs.docker.com/engine/install/linux-postinstall/#your-kernel-does-not-support-cgroup-swap-limit-capabilities
 
## Getting Started
- Install `Docker` from https://docker.io to your local machine. (Macbook is preffered)
- Talk to [@BotFather](https://t.me/BotFather) to create new your `telegram bot` and get your `BOT_TOKEN`
- Setup your instance of your cloud service provider and get ip address and ssh private key Or you can generate ssh-key pair by command like this `$ ssh-keygen -t rsa -b 4096 -f ssh_key` (no pass is preffered).
- Clone repository `git clone https://github.com/SwingbyProtocol/node-installer` and `cd node-installer`
- Store your ssh private key to accesss instance into `data/ssh_key` 
- Run `$ chmod 600 data/ssh_key` to set permission `600` to your ssh private key `data/ssh_key`
- Run `$ export BOT_TOKEN={your bot token}`
- Run `$ chmod +x scripts/start_build_and_install.sh && scripts/start_build_and_install.sh` for Mac User
- Talk to your `telegram bot` with `/start` command to start install node into your server.

## Development 
```
$ export BOT_TOKEN={your bot token}
$ go run main.go
```

## Build
```
$ make build
```

## Build docker image
```
$ make docker 
```
