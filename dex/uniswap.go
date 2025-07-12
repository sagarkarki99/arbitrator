package dex

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/constants"
	"github.com/sagarkarki99/arbitrator/contracts"
	"github.com/sagarkarki99/arbitrator/keychain"
)

func NewUniswapV3Pool(cl *ethclient.Client, kc keychain.Keychain) Dex {
	return &UniswapV3{
		cl:          cl,
		sub:         make(map[string]chan *Price),
		platformFee: 0.003, // 0.3% fee for Uniswap V3
		kc:          kc,
	}
}

type UniswapV3 struct {
	cl          *ethclient.Client
	sub         map[string]chan *Price
	platformFee float64
	kc          keychain.Keychain
}

func (u *UniswapV3) GetPrice(symbol string) (<-chan *Price, error) {
	if u.sub[symbol] != nil {
		return u.sub[symbol], nil
	}

	config, err := GetActiveMarkets(symbol, Uniswap)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool config: %w", err)
	}

	poolAddress := common.HexToAddress(config.Address)
	pool, err := contracts.NewUniswapV3Pool(poolAddress, u.cl)
	if err != nil {
		slog.Error("Could not create uniswapv3pool")
		return nil, fmt.Errorf("failed to create pool contract: %w", err)
	}

	swapChan := make(chan *contracts.UniswapV3PoolSwap)
	priceChan := make(chan *Price)

	sub, err := pool.WatchSwap(&bind.WatchOpts{}, swapChan, nil, nil)

	if err != nil {
		slog.Error("Failed to subscribe to swap events", "error", err)
		return nil, fmt.Errorf("failed to get price: %w", err)
	}
	u.sub[symbol] = priceChan
	slog.Info("Subscribed to Uniswap V3 pool", "symbol", symbol, "address", poolAddress.Hex())
	go func() {
		defer func() {
			slog.Info("Unsubscribing from Uniswap V3 pool", "symbol", symbol)
			sub.Unsubscribe()
			close(priceChan)
			delete(u.sub, symbol)
		}()
		// for i := 0; i < 3; i++ {
		// 	price := (1000 + rand.Float64()*100)
		// 	time.Sleep(time.Second * 1)
		// 	priceChan <- &Price{
		// 		Pool:            "Uniswap",
		// 		Symbol:          symbol,
		// 		Price:           price,
		// 		Liquidity:       big.NewInt(int64(price * 1e18)), // Assuming price is in USD, convert to wei
		// 		LiquidityStatus: "high",
		// 	}
		// }
		for {
			select {
			case err := <-sub.Err():
				slog.Error("Subscription error", "error", err)
				return
			case swapEvent := <-swapChan:

				price := CalculatePrice(swapEvent.SqrtPriceX96, config, symbol)
				priceChan <- &Price{
					Pool:      "Uniswap",
					Symbol:    symbol,
					Price:     price,
					Liquidity: swapEvent.Liquidity,
				}
			}
		}
	}()
	return priceChan, err
}

// Helper function to perform swaps with common logic
func (u *UniswapV3) performSwap(amount float64, symbol string, zeroForOne bool) (string, error) {
	// Step 1: Get pool configuration using the new system
	config, err := GetActiveMarkets(symbol, Uniswap)
	if err != nil {
		return "", fmt.Errorf("failed to get pool config: %w", err)
	}

	myAddress := common.HexToAddress(keychain.Accounts[0])

	// Get nonce for transaction
	nonce, err := u.cl.PendingNonceAt(context.Background(), myAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Step 5: Calculate amount with proper decimals
	var decimals int
	var tokenIn, tokenOut common.Address
	if zeroForOne {
		// Selling token0 for token1
		decimals = config.Token0Decimals
		tokenIn = common.HexToAddress(config.Token0Contract)
		tokenOut = common.HexToAddress(config.Token1Contract)
	} else {
		// Buying token0 with token1
		decimals = config.Token1Decimals
		tokenIn = common.HexToAddress(config.Token1Contract)
		tokenOut = common.HexToAddress(config.Token0Contract)
	}

	decimalMultiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	amountFloat := new(big.Float).SetFloat64(amount)
	amountWithDecimals := new(big.Float).Mul(amountFloat, decimalMultiplier)

	amountInDecimals := new(big.Int)
	amountWithDecimals.Int(amountInDecimals) // Convert to big.Int

	auth := &bind.TransactOpts{
		From:      myAddress,
		Nonce:     big.NewInt(int64(nonce)),
		Value:     big.NewInt(0), // No ETH sent with swap
		GasFeeCap: constants.GasFeeCap,
		GasTipCap: constants.GasTipCap,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return u.kc.Sign(tx)
		},
	}

	slog.Info("Swap transaction Info",
		"symbol", symbol,
		"amount", amount,
		"amount_in_decimals", amountInDecimals.String(),
		"zero_for_one", zeroForOne)

	tm := time.Now()
	swapRouter, _ := contracts.NewSwapRouter(common.HexToAddress(UniswapRouter), u.cl)
	params := contracts.IV3SwapRouterExactInputSingleParams{
		TokenIn:           tokenIn,
		TokenOut:          tokenOut,
		Recipient:         myAddress,
		AmountIn:          amountInDecimals, // Amount to swap (exact input)
		Fee:               big.NewInt(3000),
		SqrtPriceLimitX96: big.NewInt(0),
		AmountOutMinimum:  big.NewInt(0), // No minimum output amount
	}
	tx, err := swapRouter.ExactInputSingle(auth, params)

	elasped := time.Since(tm)
	if err != nil {
		slog.Error("Failed to execute swap", "error", err)
		return "", fmt.Errorf("failed to execute swap: %w", err)
	}

	slog.Info("Swap transaction submitted",
		"hash", tx.Hash().Hex(),
		"symbol", symbol,
		"amount", amount,
		"from_token", strings.Split(symbol, "/")[0],
		"to_token", strings.Split(symbol, "/")[1],
		"zero_for_one", zeroForOne,
		"executed at", elasped.String(),
	)
	return tx.Hash().Hex(), nil
}

func (u *UniswapV3) Buy(amount float64, symbol string) (string, error) {
	// Buy means: swap token1 (e.g., USDC) for token0 (e.g., WBNB)
	// zeroForOne = false (token1 → token0)
	return u.performSwap(amount, symbol, false)
}

func (u *UniswapV3) Sell(amount float64, symbol string) (string, error) {
	// Sell means: swap token0 (e.g., WBNB) for token1 (e.g., USDC)
	// zeroForOne = true (token0 → token1)
	return u.performSwap(amount, symbol, true)
}

// NOTE: The old createTransaction function has been removed and replaced with
// the new performSwap implementation above. The new implementation:
// 1. Uses the proper Uniswap V3 pool.Swap() method instead of basic transfers
// 2. Handles token decimals correctly based on pool configuration
// 3. Properly sets up transaction parameters (gas, nonce, etc.)
// 4. Provides better error handling and logging
// 5. Supports both buy and sell operations with proper swap directions

func (u *UniswapV3) GetPoolFee() float64 {
	return u.platformFee
}
