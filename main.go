package main

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/dex"
	"github.com/sagarkarki99/arbitrator/services"
)

func init() {
	// Load .env before any package-level variables are initialized
	godotenv.Load(".env")
}

func main() {

	fmt.Println("Hello arbitrator")

	cl := blockchain.Connect(blockchain.GetChains()["BscMainnetInfura"])
	defer func() {
		slog.Info("Closing client connection")
		cl.Close()
	}()

	uniswap := dex.NewUniswapV3Pool(cl)
	pancake := dex.NewPancakeswapV2Pool(cl)

	arbService := services.NewArbService(uniswap, pancake)
	arbService.Start()

}
