package dex

import (
	"fmt"
	"log/slog"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/contracts"
)

type Dex interface {
	GetPrice(symbol string) (<-chan *Price, error)
	GetPoolFee() float64
	CreateTransaction(amount float64, symbol string, from string) (string, error)
}

func NewUniswapV3Pool(cl *ethclient.Client) Dex {
	return &UniswapV3{
		cl:          cl,
		sub:         make(map[string]chan *Price),
		platformFee: 0.003, // 0.3% fee for Uniswap V3
	}
}

type UniswapV3 struct {
	cl          *ethclient.Client
	sub         map[string]chan *Price
	platformFee float64
}

func (u *UniswapV3) GetPrice(symbol string) (<-chan *Price, error) {
	config := BnbUniswapv3SymbolToPool[symbol]
	if u.sub[symbol] != nil {
		return u.sub[symbol], nil
	}
	var addr string
	if blockchain.Network == blockchain.Mainnet {
		addr = config.Address
	} else {
		addr = config.TestAddress
	}
	poolAddress := common.HexToAddress(addr)
	pool, err := contracts.NewUniswapV3Pool(poolAddress, u.cl)
	if err != nil {
		slog.Error("Could not create uniswapv3pool")
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

func (u *UniswapV3) CreateTransaction(amount float64, symbol string, from string) (string, error) {
	return "", nil
}

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
	decimalAdjustment := math.Pow(10, float64(config.token0Decimals-config.token1Decimals))
	adjustedPrice := priceFloat64 * decimalAdjustment

	// sqrtPriceX96 gives us token1/token0 ratio
	// For your pools: WETH/USDT ratio (since WETH=token1, USDT=token0)

	slog.Debug("Price calculation",
		"token0", config.token0,
		"token1", config.token1,
		"rawPrice", priceFloat64,
		"adjustedPrice", adjustedPrice,
		"desiredPair", desiredPair)

	// adjustedPrice is WETH/USDT (token1/token0)
	// If desired pair is "WETH/USDT" or "ETH/USDT", return as-is (USDT per WETH)
	// If desired pair were "USDT/WETH", we'd return 1/adjustedPrice

	return adjustedPrice // This gives USDT per WETH
}

// {
//     "timestamp": "2025/06/24 11:31:48",
//     "level": "INFO",
//     "message": "New swap event received",
//     "sender": "0x66a9893cc07d91d95644aedd05d03f95e1dba8af",
//     "amount0": 731413563172656263,
//     "amount1": -1757750507,
//     "sqrtPriceX96": 3889799402949823877143858,
//     "tick": -198445,
//     "calculated_price": 3436.90813798861
// }
