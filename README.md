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
```
$ git clone https://github.com/SwingbyProtocol/node-installer && cd node-installer
```
Install steps (Let's do on your local machine)
- Install `Docker` from https://docker.io to your local machine. (Macbook is preffered)
- Talk to [@BotFather](https://t.me/BotFather) to create new your `telegram bot` and get your `BOT_TOKEN`
- Setup your instance of your cloud service provider
- Get IP address (v4) and SSH private key
- OR => create SSH key pair `$ ssh-keygen -t rsa -b 4096 -f ssh_key` (no passphrase).
- Store your ssh private key to accesss instance into `data/ssh_key` 
- `$ chmod 600 data/ssh_key` to set permission `600` to your ssh private key `data/ssh_key`
- `$ export BOT_TOKEN={your bot token}`
- `$ chmod +x scripts/start_build_and_install.sh && scripts/start_build_and_install.sh` for Mac User
- Talk to your `telegram bot` with `/start` command to moving bot into your server.

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
