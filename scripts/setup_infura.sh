#!/bin/bash
if ! type "docker" >/dev/null; then
    apt-get update
    apt-get install -y docker.io git
    service docker start
fi
if ! type "docker-compose" >/dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/download/1.27.4/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi
