#!/bin/bash

docker pull swingbylabs/node-installer:latest

docker run -v $(pwd)/data:/app/data --name node_installer -e BOT_TOKEN --rm swingbylabs/node-installer:latest
