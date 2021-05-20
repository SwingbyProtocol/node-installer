build-linux-amd64:
		GOOS=linux GOARCH=amd64 go build -o bin/bot_linux_amd64 . 
build:
		go build -o bin/bot .

docker:
		docker build -t swingbylabs/node-installer:latest .
push:
		docker push swingbylabs/node-installer:latest
release:
		docker tag swingbylabs/node-installer:latest swingbylabs/node-installer:1.1.10
		docker push swingbylabs/node-installer:1.1.10