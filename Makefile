build-linux-amd64:
		GOOS=linux GOARCH=amd64 go build -o bin/bot_linux_amd64 .
build:
		go build -o bin/bot .