package dex

import (
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/contracts"

	"github.com/sagarkarki99/arbitrator/keychain"
)

func NewPancakeswapV2Pool(client *ethclient.Client, kc keychain.Keychain) Dex {
	pool := &PancakeswapV2Pool{
		cl:          client,
		subs:        make(map[string]chan *Price),
		platformFee: 0.002, // 0.2% platform fee
		kc:          kc,
	}

	return pool
}

type PancakeswapV2Pool struct {
	cl          *ethclient.Client
	subs        map[string]chan *Price
	platformFee float64
	kc          keychain.Keychain
}

func (p *PancakeswapV2Pool) GetPrice(symbol string) (<-chan *Price, error) {
	if sub, exists := p.subs[symbol]; exists {
		return sub, nil
	}

	config, err := GetActiveMarkets(symbol, Pancakeswap)
	if err != nil {
		return nil, fmt.Errorf("no pool found for symbol %s", symbol)
	}

	poolAddress := common.HexToAddress(config.Address)
	pool, err := contracts.NewPancakeswapV3Pool(poolAddress, p.cl)
	if err != nil {
		return nil, err
	}

	s, _ := pool.ProtocolFees(nil)
	slog.Info("Protocol fee", s.Token0, s.Token1)

	swapChan := make(chan *contracts.PancakeswapV3PoolSwap)
	priceChan := make(chan *Price)

	sub, err := pool.WatchSwap(&bind.WatchOpts{}, swapChan, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not subscribe to swap events: %w", err)
	}

	p.subs[symbol] = priceChan

	go func() {
		defer func() {
			sub.Unsubscribe()
			close(priceChan)
			delete(p.subs, symbol)
			slog.Info("Unsubscribing from Pancakeswap V2 pool", "symbol", symbol, "address", config.Address)
		}()
		// for i := 0; i < 3; i++ {
		// 	price := (1000 + rand.Float64()*5) // Random price between 1000-1100
		// 	time.Sleep(time.Second * 2)
		// 	priceChan <- &Price{
		// 		Pool:            "Pancakeswap",
		// 		Symbol:          symbol,
		// 		Price:           price,
		// 		Liquidity:       big.NewInt(1000000), // Example liquidity value
		// 		LiquidityStatus: "high",
		// 	}
		// }
		// close(priceChan)
		for {
			select {
			case err := <-sub.Err():
				slog.Error("Error in Pancakeswap V2 pool subscription", "error", err)
				return
			case swapEvent := <-swapChan:
				price := CalculatePrice(swapEvent.SqrtPriceX96, config, symbol)
				status := "high"
				if swapEvent.Liquidity.Cmp(big.NewInt(1e10)) < 0 {
					status = "low"
				}
				priceChan <- &Price{
					Pool:            "Pancakeswap",
					Symbol:          symbol,
					Price:           price,
					Liquidity:       swapEvent.Liquidity,
					LiquidityStatus: status,
				}
			}
		}
	}()
	return priceChan, nil
}

func (p *PancakeswapV2Pool) Buy(amount float64, symbol string) (string, error) {
	return p.performSwap(amount, symbol, false)
}

func (p *PancakeswapV2Pool) Sell(amount float64, symbol string) (string, error) {
	return p.performSwap(amount, symbol, true)
}

// Helper function to perform swaps with common logic
func (p *PancakeswapV2Pool) performSwap(amount float64, symbol string, zeroForOne bool) (string, error) {
	// Step 1: Get pool configuration
	config, err := GetActiveMarkets(symbol, Pancakeswap)
	if err != nil {
		return "", fmt.Errorf("pool configuration not found for symbol: %s", symbol)
	}

	// Step 3: Initialize pool contract
	poolAddress := common.HexToAddress(config.Address)
	pool, err := contracts.NewPancakeswapV3Pool(poolAddress, p.cl)
	if err != nil {
		slog.Error("Could not create pancakeswap pool", "error", err)
		return "", fmt.Errorf("failed to create pool contract: %w", err)
	}

	// Step 4: Set up transaction parameters
	myAddress := common.HexToAddress(keychain.Accounts[0])

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

	// Step 6: Prepare transaction options
	auth := &bind.TransactOpts{
		From:      myAddress,
		Nonce:     big.NewInt(int64(15)),
		Value:     big.NewInt(0), // No ETH value for token swaps
		GasFeeCap: TestGasFeeCap,
		GasTipCap: TestGasTipCap,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return p.kc.Sign(tx)
		},
	}

	tm := time.Now()
	// Step 7: Execute the swap
	tx, err := pool.Swap(
		auth,
		myAddress,        // recipient
		zeroForOne,       // swap direction
		amountInDecimals, // amountSpecified (exact input)
		big.NewInt(0),    // sqrtPriceLimitX96 (no limit)
		[]byte{},         // data (empty)
	)

	elasped := time.Since(tm)
	slog.Info("Swap execution time", "duration", elasped)

	if err != nil {
		if err != nil {
			slog.Error("Swap failed with details",
				"error", err,
				"symbol", symbol,
				"amount", amount,
				"nonce", 18,
				"pool_address", poolAddress.Hex())
			return "", fmt.Errorf("failed to execute swap: %w", err)
		}
	}

	slog.Info("Swap transaction submitted",
		"hash", tx.Hash().Hex(),
		"symbol", symbol,
		"amount", amount,
		"from_token", fromToken,
		"to_token", toToken,
		"zero_for_one", zeroForOne)

	return tx.Hash().Hex(), nil
}

func (p *PancakeswapV2Pool) GetPoolFee() float64 {
	return p.platformFee
}
