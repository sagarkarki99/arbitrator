package dex

import (
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/contracts"
)

var Uniswapv3SymbolToPool = map[string]string{
	"ETH/USDT": "0x11b815efB8f581194ae79006d24E0d814B7697F6",
}

type Price struct {
	Symbol string
	Price  float64
}

type Dex interface {
	GetPrice(symbol string) (<-chan *Price, error)
	CreateTransaction(amount float64, symbol string, from string) (string, error)
}

func NewUniswapV3Pool(cl *ethclient.Client) Dex {
	return &UniswapV3{
		cl:  cl,
		sub: make(map[string]chan *Price),
	}
}

type UniswapV3 struct {
	cl  *ethclient.Client
	sub map[string]chan *Price
}

func (u *UniswapV3) GetPrice(symbol string) (<-chan *Price, error) {
	poolAddr := Uniswapv3SymbolToPool[symbol]
	if u.sub[symbol] != nil {
		return u.sub[symbol], nil
	}
	poolAddress := common.HexToAddress(poolAddr)
	pool, err := contracts.NewUniswapV3Pool(poolAddress, u.cl)
	if err != nil {
		slog.Error("Could not create uniswapv3pool")
	}

	swapChan := make(chan *contracts.UniswapV3PoolSwap)
	priceChan := make(chan *Price)

	sub, err := pool.WatchSwap(&bind.WatchOpts{}, swapChan, nil, nil)

	if err != nil {
		slog.Error("Failed to subscribe to swap events", "error", err)
		return nil, fmt.Errorf("Failed to get price: %w", err)
	}
	u.sub[symbol] = priceChan

	go func() {
		for {
			select {
			case err := <-sub.Err():
				slog.Error("Subscription error", "error", err)
				sub.Unsubscribe()
				return
			case swapEvent := <-swapChan:

				slog.Info("New swap event received",
					"sender", swapEvent.Sender,
					"amount0", swapEvent.Amount0,
					"amount1", swapEvent.Amount1,
					"sqrtPriceX96", swapEvent.SqrtPriceX96,
					"tick", swapEvent.Tick,
					"calculated_price", 4.5)

				priceChan <- &Price{
					Symbol: symbol,
					Price:  4.5,
				}
			}
		}
	}()
	return priceChan, err
}

func (u *UniswapV3) CreateTransaction(amount float64, symbol string, from string) (string, error) {
	return "", nil
}
