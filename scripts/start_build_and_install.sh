#!/bin/bash

#docker build -t swingbylabs/node-installer .

#docker push swingbylabs/node-installer

docker run -v $(pwd)/data:/app/data --name node_installer -e BOT_TOKEN --rm swingbylabs/node-installer:1.0.0
