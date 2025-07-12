package main

import (
	"fmt"
	"log/slog"

	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
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
	// kc := keychain.NewKeychainImpl()

	// Check token balance before swapping
	config, _ := dex.GetActiveMarkets("USDC/WETH", dex.Uniswap)
	fmt.Printf("Pool config: %+v\n", config)

	// if !IsEnoughBalanceApproved(config, cl) {
	// 	keychain.SendApproval(cl, "0x3bFA4769FB09eefC5a80d6E87c3B9C650f7Ae48E", config.Token1Contract, kc)
	// 	slog.Info("Approval transaction sent, waiting for confirmation...")
	// }

	// uniswap := dex.NewUniswapV3Pool(cl, kc)
	// pancake := dex.NewPancakeswapV2Pool(cl, kc)

	keychain.GetNativeBalance(keychain.Accounts[0], cl)
	bal, _ := keychain.GetBalance(keychain.Accounts[0], config.Token0Contract, cl)
	b := keychain.ConvertReadable(bal, config.Token0Decimals)
	fmt.Printf(" Balance: %s %s\n", b.String(), config.Token0)

	// config, _ := dex.GetActiveMarkets("USDC/ETH", dex.Uniswap)
	// keychain.SendApproval(cl, config.Address, config.Token0Contract, kc)

	// arbService := services.NewArbService(uniswap, pancake)
	// arbService.Start("USDC/ETH")

}

func IsEnoughBalanceApproved(config *dex.PoolConfig, cl *ethclient.Client) bool {
	balance, err := keychain.GetBalance(keychain.Accounts[0], config.Token0Contract, cl)
	if err != nil {
		slog.Error("Failed to get balance", "error", err)
		return true
	}

	bal := keychain.ConvertReadable(balance, config.Token0Decimals)
	fmt.Printf("USDC Balance (human readable): %s\n", bal.String())

	if balance.Cmp(big.NewInt(0)) == 0 {
		slog.Error("You have 0 USDC tokens! Cannot perform swap.")
		return false
	}

	balance, err = keychain.GetBalance(keychain.Accounts[0], config.Token1Contract, cl)
	if err != nil {
		slog.Error("Failed to convert balance", "error", err)
		return false
	}

	balInEth := keychain.ConvertReadable(balance, config.Token1Decimals)

	fmt.Printf("WETH Balance: %s (raw: %s)\n", balInEth.String(), bal.String())

	keychain.GetNativeBalance(keychain.Accounts[0], cl)

	// Check approval for the pool
	return keychain.HasApproval(config.Address, cl)

}

/*
1. Select a pool to swap tokens, starting balance. eg: WBTC/WETH, 0.01 ETH
2. Approve the pool/swapRouter of both dexs/tokenpool to spend the tokens. This is to make the swap smooth
 For eg: swap WBTC for WETH in shushiswap (shushiswap router should approved for to use WETH token)
         and swap WBTC for WETH in pancakeswap (pancakeswap router should approved for to use WBTC token.)
3. This approval is done before starting to arbitrage.

*/
