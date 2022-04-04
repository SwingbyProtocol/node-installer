module github.com/SwingbyProtocol/node-installer

go 1.14

require (
	github.com/SwingbyProtocol/tx-indexer v0.0.0-20200809124002-e54d6740619f
	github.com/apenella/go-ansible v0.7.1
	github.com/binance-chain/go-sdk v1.2.6
	github.com/binance-chain/tss-lib v1.3.2
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/perlin-network/noise v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	google.golang.org/genproto v0.0.0-20210212180131-e7f2df4ecc2d // indirect
	google.golang.org/grpc v1.35.0 // indirect
)

replace github.com/perlin-network/noise => github.com/SwingbyProtocol/noise v1.1.1-0.20200203090125-898aaedd390e

replace github.com/binance-chain/tss-lib => gitlab.com/thorchain/tss/tss-lib v0.0.0-20200809185403-362e7ed851e4
