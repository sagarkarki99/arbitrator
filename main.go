package main

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/dex"
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

	pool := dex.NewUniswapV3Pool(cl)
	priceChan, err := pool.GetPrice("WETH/USDT")
	if err != nil {
		slog.Error("Failed to get price", "error", err)
		return
	}

	for {
		price := <-priceChan
		slog.Info("Received price update", "symbol", price.Symbol, "price", price.Price)
	}

}
