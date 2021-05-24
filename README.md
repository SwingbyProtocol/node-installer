# node-installer
[![Go Report Card](https://goreportcard.com/badge/github.com/SwingbyProtocol/node-installer)](https://goreportcard.com/report/github.com/SwingbyProtocol/node-installer)

## Requirements (Mainnet environment)

- OS Ubuntu 20.04 LTS
- CPUs >= 4
- Memory >= 16GB
- Disk space on `/var/swingby` >= 1.5TB:ETH, 1.6TB:BSC
- Swap memory >= 5GB
- Network bandwidth >= 500Mbps

The docker host should supports swap memory. [details](https://docs.docker.com/config/containers/resource_constraints/)

To enable swap memory, set up your configuration and restart your instance according to this document.
https://docs.docker.com/engine/install/linux-postinstall/#your-kernel-does-not-support-cgroup-swap-limit-capabilities

Note: `ufw` should be disabled for running Swingby node.
 
## Getting Started
```
$ git clone https://github.com/SwingbyProtocol/node-installer && cd node-installer
```
Install steps (Let's execute on your local machine)
1. Install `Docker` from https://docker.io to your local machine. (Macbook is preffered)
2. Talk to [@BotFather](https://t.me/BotFather) to create new your `telegram bot` and get your `BOT_TOKEN`
3. Setup your instance on the cloud service provider. (_note: if you haven't SSH key, you can create SSH key pair_)
```
$ ssh-keygen -t rsa -b 4096 -f ssh_key   // (no passphrase)
```
4. Get IP address (v4) and SSH private key for your server.
5. Store your SSH private key into `data/ssh_key` 
6. Set permission `600` to `data/ssh_key` file
```bash
$ chmod 600 data/ssh_key
```
7. Set env variable

```bash
$ export BOT_TOKEN={your bot token}
```
8. Run Bot on your local machine.
```bash
$ chmod +x scripts/install.sh && scripts/install.sh
```
9. Talk to your `telegram bot` with `/start` command to moving bot into your server.

## Development 
```golang
$ BOT_TOKEN={your bot token} go run main.go
```

## Build
```
$ make build
```

## Build docker image
```
$ make docker 
```
