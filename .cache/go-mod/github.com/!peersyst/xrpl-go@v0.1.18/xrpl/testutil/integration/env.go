package integration

import (
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
)

// EnvKey is the key for the integration environment.
type EnvKey string

const (
	// LocalnetEnv is the environment key for a local XRPL network.
	LocalnetEnv EnvKey = "localnet"
	// TestnetEnv is the environment key for the public testnet.
	TestnetEnv EnvKey = "testnet"
	// DevnetEnv is the environment key for the developer network.
	DevnetEnv EnvKey = "devnet"
)

// IntegrationWebsocketEnvs is the map of websocket integration environments.
var IntegrationWebsocketEnvs = map[EnvKey]Env{
	LocalnetEnv: {
		Host:           "ws://0.0.0.0:6006",
		FaucetProvider: nil,
	},
	TestnetEnv: {
		Host:           "wss://s.altnet.rippletest.net:51233",
		FaucetProvider: faucet.NewTestnetFaucetProvider(),
	},
	DevnetEnv: {
		Host:           "wss://s.devnet.rippletest.net:51233",
		FaucetProvider: faucet.NewDevnetFaucetProvider(),
	},
}

// IntegrationRPCEnvs is the map of RPC integration environments.
var IntegrationRPCEnvs = map[EnvKey]Env{
	LocalnetEnv: {
		Host:           "http://0.0.0.0:5005",
		FaucetProvider: nil,
	},
	TestnetEnv: {
		Host:           "https://s.altnet.rippletest.net:51234",
		FaucetProvider: faucet.NewTestnetFaucetProvider(),
	},
	DevnetEnv: {
		Host:           "https://s.devnet.rippletest.net:51234",
		FaucetProvider: faucet.NewDevnetFaucetProvider(),
	},
}

// Env is the environment for the integration tests.
// It contains the host and the faucet provider.
type Env struct {
	Host           string
	FaucetProvider FaucetProvider
}
