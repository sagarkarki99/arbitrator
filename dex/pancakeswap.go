package dex

import (
	"fmt"
	"log/slog"
	"math/big"
	"math/rand/v2"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/contracts"
	"github.com/sagarkarki99/arbitrator/keychain"
)

func NewPancakeswapV2Pool(client *ethclient.Client) Dex {
	pool := &PancakeswapV2Pool{
		cl:          client,
		subs:        make(map[string]chan *Price),
		platformFee: 0.002, // 0.2% platform fee
	}

	return pool
}

type PancakeswapV2Pool struct {
	cl          *ethclient.Client
	subs        map[string]chan *Price
	platformFee float64
}

func (p *PancakeswapV2Pool) GetPrice(symbol string) (<-chan *Price, error) {
	if sub, exists := p.subs[symbol]; exists {
		return sub, nil
	}

	config, exists := BnbPancakeTestnetSymbolToPool[symbol]
	if !exists {
		return nil, fmt.Errorf("no pool found for symbol %s", symbol)
	}

	var addr string
	if blockchain.ActiveChain.Network == "mainnet" {
		addr = config.Address
	} else {
		addr = config.TestAddress
	}
	poolAddress := common.HexToAddress(addr)
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
		for i := 0; i < 3; i++ {
			price := (1000 + rand.Float64()*5) // Random price between 1000-1100
			time.Sleep(time.Second * 2)
			priceChan <- &Price{
				Pool:            "Uniswap",
				Symbol:          symbol,
				Price:           price,
				Liquidity:       big.NewInt(1000000), // Example liquidity value
				LiquidityStatus: "high",
			}
		}
		close(priceChan)
		// for {
		// 	select {
		// 	case err := <-sub.Err():
		// 		slog.Error("Error in Pancakeswap V2 pool subscription", "error", err)
		// 		return
		// 	case swapEvent := <-swapChan:
		// 		// price := CalculatePrice(swapEvent.SqrtPriceX96, config, symbol)
		// 		// status := "high"
		// 		// if swapEvent.Liquidity.Cmp(big.NewInt(1e10)) < 0 {
		// 		// 	status = "low"
		// 		// }
		// 		priceChan <- &Price{
		// 			Pool:            "Uniswap",
		// 			Symbol:          symbol,
		// 			Price:           price,
		// 			Liquidity:       swapEvent.Liquidity,
		// 			LiquidityStatus: status,
		// 		}
		// 	}
		// }
	}()
	return priceChan, nil
}

func (p *PancakeswapV2Pool) Buy(amount float64) (string, error) {
	return p.createTransaction(amount, keychain.Accounts[0])
}

func (p *PancakeswapV2Pool) Sell(amount float64) (string, error) {
	return p.createTransaction(amount, keychain.Accounts[0])
}

func (p *PancakeswapV2Pool) createTransaction(amount float64, from string) (string, error) {
	// Implement transaction creation logic here
	return "", nil
}

func (p *PancakeswapV2Pool) GetPoolFee() float64 {
	return p.platformFee
}
