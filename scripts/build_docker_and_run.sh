#!/bin/bash

make docker && make push

docker run -v $(pwd)/data:/app/data --name node_installer -e BOT_TOKEN --rm swingbylabs/node-installer:1.0.0
