#!/bin/bash

docker build -t swingbylabs/node-installer . 

docker run -v $(PWD)/data:/app/data --name node-installer -e BOT_TOKEN --rm swingbylabs/node-installer
