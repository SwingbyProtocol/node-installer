module github.com/SwingbyProtocol/node-installer

go 1.14

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/SwingbyProtocol/tx-indexer v0.0.0-20200809124002-e54d6740619f
	github.com/apenella/go-ansible v0.5.0
	github.com/binance-chain/go-sdk v1.2.5
	github.com/binance-chain/tss-lib v1.3.2
	github.com/containerd/containerd v1.4.3 // indirect
	github.com/cosmos/go-bip39 v0.0.0-20200817134856-d632e0d11689
	github.com/docker/docker v20.10.2+incompatible // indirect
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/perlin-network/noise v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.7.0
	github.com/sparrc/containermon v0.0.0-20200519161504-2a4fa992796c // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/sys v0.0.0-20210113181707-4bcb84eeeb78 // indirect
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f
	google.golang.org/grpc v1.35.0 // indirect

)

replace github.com/perlin-network/noise => github.com/SwingbyProtocol/noise v1.1.1-0.20200203090125-898aaedd390e

replace github.com/binance-chain/tss-lib => gitlab.com/thorchain/tss/tss-lib v0.0.0-20200809185403-362e7ed851e4
