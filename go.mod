module github.com/SwingbyProtocol/node-installer

go 1.14

require (
	github.com/apenella/go-ansible v0.5.0
	github.com/binance-chain/go-sdk v1.2.5
	github.com/binance-chain/tss-lib v1.3.2
	github.com/cosmos/go-bip39 v0.0.0-20200817134856-d632e0d11689
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/perlin-network/noise v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.7.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de

)

replace github.com/perlin-network/noise => github.com/SwingbyProtocol/noise v1.1.1-0.20200203090125-898aaedd390e

replace github.com/binance-chain/tss-lib => gitlab.com/thorchain/tss/tss-lib v0.0.0-20200809185403-362e7ed851e4
