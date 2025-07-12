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
	"log/slog"
	"math"
	"math/big"
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

var (
	UniswapRouter     = "0x3bFA4769FB09eefC5a80d6E87c3B9C650f7Ae48E"
	PancakeswapRouter = ""
)

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
