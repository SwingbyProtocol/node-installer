#!/bin/bash

docker build -t install_swingby . && docker run -v $(PWD)/configs:/app/configs --name install_swingby -e BOT_TOKEN --rm install_swingby 