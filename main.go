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

	uniswap := dex.NewUniswapV3Pool(cl)
	pancake := dex.NewPancakeswapV2Pool(cl)

	uniswapPrice, err := uniswap.GetPrice("WETH/USDT")
	if err != nil {
		slog.Error("Failed to get price from UNISWAP", "error", err)
	}
	pancakePrice, err := pancake.GetPrice("WETH/USDT")
	if err != nil {
		slog.Error("Failed to get price from PANCAKE", "error", err)
	}

	for {
		select {
		case uPrice := <-uniswapPrice:
			slog.Info("Received Uniswap price update", "symbol", uPrice.Symbol, "price", uPrice.Price, "liquidity", uPrice.Liquidity, "Liquidity statu", uPrice.LiquidityStatus)
		case pPrice := <-pancakePrice:
			slog.Info("Received Pancakeswap price update", "symbol", pPrice.Symbol, "price", pPrice.Price, "liquidity", pPrice.Liquidity, "Liquidity status", pPrice.LiquidityStatus)
		}
	}

}
