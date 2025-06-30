package main

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/dex"
	"github.com/sagarkarki99/arbitrator/services"
)

func main() {
	fmt.Println("Hello arbitrator")
	godotenv.Load(
		".env",
	)

	cl := blockchain.Connect()
	defer func() {
		slog.Info("Closing Ethereum client connection")
		cl.Close()
	}()

	uniswap := dex.NewUniswapV3Pool(cl)
	pancake := dex.NewPancakeswapV2Pool(cl)

	arbService := services.NewArbService(uniswap, pancake)
	arbService.Start()

}
