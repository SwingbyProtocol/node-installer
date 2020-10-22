#!/bin/bash

docker build -t install_swingby . && docker run --name install_swingby -e BOT_TOKEN --rm install_swingby 