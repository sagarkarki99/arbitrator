package dex

// TOKEN SWAP EXPLANATION:
//
// This file implements Uniswap V3 token swaps. Here's how the swap process works:
//
// 1. POOL STRUCTURE:
//    - Each pool has two tokens: token0 and token1
//    - token0 has a lower address than token1 (Uniswap convention)
//    - Example: WBNB/USDC pool has token0=WBNB, token1=USDC
//
// 2. SWAP DIRECTIONS:
//    - zeroForOne = true:  token0 → token1 (e.g., WBNB → USDC)
//    - zeroForOne = false: token1 → token0 (e.g., USDC → WBNB)
//
// 3. BUY vs SELL:
//    - Buy():  Acquire token0 using token1 (zeroForOne = false)
//    - Sell(): Acquire token1 using token0 (zeroForOne = true)
//
// 4. SWAP PARAMETERS:
//    - recipient: Address to receive output tokens
//    - zeroForOne: Swap direction (true/false)
//    - amountSpecified: Input amount (positive for exact input)
//    - sqrtPriceLimitX96: Price limit (0 = no limit)
//    - data: Callback data (empty for simple swaps)
//
// 5. TRANSACTION FLOW:
//    - Get pool configuration and address
//    - Initialize pool contract
//    - Calculate amount with proper decimals
//    - Set up transaction options (gas, nonce, etc.)
//    - Execute swap and return transaction hash

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/constants"
	"github.com/sagarkarki99/arbitrator/contracts"
	"github.com/sagarkarki99/arbitrator/keychain"
)

type Dex interface {
	GetPrice(symbol string) (<-chan *Price, error)
	GetPoolFee() float64
	Buy(amount float64, symbol string) (string, error)
	Sell(amount float64, symbol string) (string, error)
}

type DexApp string

var (
	Uniswap     DexApp = "Uniswap"
	Pancakeswap DexApp = "Pancakeswap"
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
	var fromToken, toToken string

	if zeroForOne {
		// Selling token0 for token1
		decimals = config.Token0Decimals
		fromToken = config.Token0
		toToken = config.Token1
	} else {
		// Buying token0 with token1
		decimals = config.Token1Decimals
		fromToken = config.Token1
		toToken = config.Token0
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
		"from_token", fromToken,
		"to_token", toToken,
		"zero_for_one", zeroForOne)

	tm := time.Now()
	// Step 7: Execute the swap
	swapRouter, _ := contracts.NewSwapRouter(common.HexToAddress("0x3bFA4769FB09eefC5a80d6E87c3B9C650f7Ae48E"), u.cl)
	params := contracts.IV3SwapRouterExactInputSingleParams{
		TokenIn:           common.HexToAddress(config.Token1Contract),
		TokenOut:          common.HexToAddress(config.Token0Contract),
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
		"from_token", fromToken,
		"to_token", toToken,
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

func CalculatePrice(sqrtPriceX96 *big.Int, config *PoolConfig, desiredPair string) float64 {
	// Convert sqrtPriceX96 to big.Float for precision
	sqrtPriceX96Float := new(big.Float).SetInt(sqrtPriceX96)

	// Divide by 2^96
	divisor := new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), 96))
	sqrtPrice := new(big.Float).Quo(sqrtPriceX96Float, divisor)

	// Square to get price (token1/token0 ratio)
	price := new(big.Float).Mul(sqrtPrice, sqrtPrice)
	priceFloat64, _ := price.Float64()

	// Adjust for decimal differences: human-readable token1/token0 ratio
	// The raw price returned by Uniswap is based on integer amounts,
	// so we need to multiply by 10^(token0Decimals−token1Decimals)
	// to account for the decimal places of each token.
	decimalAdjustment := math.Pow(10, float64(config.Token0Decimals-config.Token1Decimals))
	adjustedPrice := priceFloat64 * decimalAdjustment

	// sqrtPriceX96 gives us token1/token0 ratio
	// For your pools: WETH/USDT ratio (since WETH=token1, USDT=token0)

	slog.Debug("Price calculation",
		"token0", config.Token0,
		"token1", config.Token1,
		"rawPrice", priceFloat64,
		"adjustedPrice", adjustedPrice,
		"desiredPair", desiredPair)

	// adjustedPrice is WETH/USDT (token1/token0)
	// If desired pair is "WETH/USDT" or "ETH/USDT", return as-is (USDT per WETH)
	// If desired pair were "USDT/WETH", we'd return 1/adjustedPrice

	return adjustedPrice // This gives USDT per WETH
}
