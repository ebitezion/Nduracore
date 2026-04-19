package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	fmt.Println("⏳ Connecting to testnet...")
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.altnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("Error disconnecting:", err)
		}
	}()

	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	if !client.IsConnected() {
		fmt.Println("❌ Failed to connect to testnet")
		return
	}

	fmt.Println("✅ Connected to testnet")
	fmt.Println()

	wallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}

	balance, err := client.GetXrpBalance(wallet.ClassicAddress)
	if err != nil {
		balance = "0"
	}

	fmt.Println("💳 Balance", balance)
	fmt.Println()

	fmt.Println("⏳ Funding wallet...")
	err = client.FundWallet(&wallet)
	if err != nil {
		fmt.Println(err)
		return
	}

	balance, err = client.GetXrpBalance(wallet.ClassicAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("💳 Balance", balance)
	fmt.Println()

	fmt.Println("⏳ Funding wallet on devnet...")

	balance, err = client.GetXrpBalance(wallet.ClassicAddress)
	if err != nil {
		balance = "0"
	}

	fmt.Println("💳 Balance", balance)
	fmt.Println()

	fmt.Println("⏳ Funding wallet...")
	err = client.FundWallet(&wallet)
	if err != nil {
		fmt.Println(err)
		return
	}

	balance, err = client.GetXrpBalance(wallet.ClassicAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("💳 Balance", balance)
}
