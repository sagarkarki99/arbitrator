package dex

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"math/rand/v2"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/contracts"
	"github.com/sagarkarki99/arbitrator/keychain"
)

type Dex interface {
	GetPrice(symbol string) (<-chan *Price, error)
	GetPoolFee() float64
	Buy(amount float64) (string, error)
	Sell(amount float64) (string, error)
}

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
	config := BnbPancakeTestnetSymbolToPool[symbol]
	if u.sub[symbol] != nil {
		return u.sub[symbol], nil
	}
	var addr string
	if blockchain.ActiveChain.Network == "mainnet" {
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
		for i := 0; i < 3; i++ {
			price := (1000 + rand.Float64()*100)
			time.Sleep(time.Second * 1)
			priceChan <- &Price{
				Pool:            "Uniswap",
				Symbol:          symbol,
				Price:           price,
				Liquidity:       big.NewInt(int64(price * 1e18)), // Assuming price is in USD, convert to wei
				LiquidityStatus: "high",
			}
		}
		// for {
		// 	select {
		// 	case err := <-sub.Err():
		// 		slog.Error("Subscription error", "error", err)
		// 		return
		// 	case swapEvent := <-swapChan:

		// 		price := CalculatePrice(swapEvent.SqrtPriceX96, config, symbol)
		// 		priceChan <- &Price{
		// 			Pool:      "Uniswap",
		// 			Symbol:    symbol,
		// 			Price:     price,
		// 			Liquidity: swapEvent.Liquidity,
		// 		}
		// 	}
		// }
	}()
	return priceChan, err
}

func (u *UniswapV3) Buy(amount float64) (string, error) {
	return u.createTransaction(amount, keychain.Accounts[0])
}
func (u *UniswapV3) Sell(amount float64) (string, error) {
	return u.createTransaction(amount, keychain.Accounts[0])
}

func (u *UniswapV3) createTransaction(amount float64, to string) (string, error) {
	// config := BnbUniswapv3SymbolToPool[symbol]
	// var addr string
	// if blockchain.ActiveChain.Network == "mainnet" {
	// 	addr = config.Address
	// } else {
	// 	addr = config.TestAddress
	// }
	// poolAddress := common.HexToAddress(addr)
	// pool, err := contracts.NewUniswapV3Pool(poolAddress, u.cl)

	// //pool.swap will broadcast the swap

	receiver := common.HexToAddress(to)
	sender := common.HexToAddress(keychain.Accounts[0])
	nounce, _ := u.cl.PendingNonceAt(context.Background(), sender)

	value := big.NewInt(int64(amount * 1e18))

	slog.Info("Network ID", string(blockchain.ActiveChain.ChainID), "Nounce", nounce, "value", value)
	suggestedFee, _ := u.cl.SuggestGasPrice(context.Background())

	trx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(blockchain.ActiveChain.ChainID)),
		Nonce:     nounce,
		To:        &receiver,
		GasFeeCap: big.NewInt(300_000_000), // 0.1 gwei
		GasTipCap: big.NewInt(100_000_000), // 0.001 gwei
		Value:     value,                   // Convert amount to wei
	})

	slog.Info("Suggested gas for transaction",
		"fee", suggestedFee)

	signedTx, err := u.kc.Sign(trx)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}
	if err := u.cl.SendTransaction(context.Background(), signedTx); err != nil {
		slog.Error("Failed to send transaction", "error", err)
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}
	slog.Info("Transaction sent", "hash", signedTx.Hash().Hex())
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
