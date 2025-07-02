package main

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/dex"
	"github.com/sagarkarki99/arbitrator/keychain"
	"github.com/sagarkarki99/arbitrator/services"
)

func init() {
	// Load .env before any package-level variables are initialized
	godotenv.Load(".env")
}

func main() {

	fmt.Println("Hello arbitrator")

	cl := blockchain.Connect(blockchain.GetChains()["BscTestnet"])
	defer func() {
		slog.Info("Closing client connection")
		cl.Close()
	}()
	kc := keychain.NewKeychainImpl()
	uniswap := dex.NewUniswapV3Pool(cl, kc)
	pancake := dex.NewPancakeswapV2Pool(cl)

	// uniswap.CreateTransaction(0.001, "BNB", "0x267f5BFc9486b446046e97F9D3864793dA8eFb33")
	// time.Sleep(5 * time.Second)
	// uniswap.GetPoolFee()
	arbService := services.NewArbService(uniswap, pancake)
	arbService.Start()

}
