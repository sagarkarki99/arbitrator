package main

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/dex"
	"github.com/sagarkarki99/arbitrator/keychain"
)

func init() {
	// Load .env before any package-level variables are initialized
	godotenv.Load(".env")
}

func main() {

	fmt.Println("Hello arbitrator")

	cl := blockchain.Connect(blockchain.GetChains()["EthSepolia"])
	defer func() {
		slog.Info("Closing client connection")
		cl.Close()
	}()

	config, _ := dex.GetActiveMarkets("USDC/WETH", dex.Uniswap)
	fmt.Printf("Pool config: %+v\n", config)

	kc := keychain.NewKeychainImpl()
	uniswap := dex.NewUniswapV3Pool(cl, kc)
	// pancake := dex.NewPancakeswapV2Pool(cl, kc)
	hash, err := uniswap.Sell(10, "USDC/WETH")
	if err != nil {
		slog.Error("Failed to buy tokens", "error", err)
		return
	}

	slog.Info("Transaction hash:", "hash", hash)
	// arbService := services.NewArbService(uniswap, pancake)
	// arbService.Start("USDC/WETH")

}

/*
1. Select a pool to swap tokens, starting balance. eg: WBTC/WETH, 0.01 ETH
2. Approve the pool/swapRouter of both dexs/tokenpool to spend the tokens. This is to make the swap smooth
 For eg: swap WBTC for WETH in shushiswap (shushiswap router should approved for to use WETH token)
         and swap WBTC for WETH in pancakeswap (pancakeswap router should approved for to use WBTC token.)
3. This approval is done before starting to arbitrage.

*/
