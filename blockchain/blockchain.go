package blockchain

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/contracts"
)

type Blockchain interface {
	Connect()
}

func Connect() {
	apiKey := os.Getenv("INFURA_API_KEY")
	wsUrl := fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s", apiKey)
	cl, err := ethclient.DialContext(context.Background(), wsUrl)
	if err != nil {
		panic(err)
	}
	slog.Info("Connected to Ethereum client", "url", wsUrl)
	defer func() {
		slog.Info("Closing Ethereum client connection")
		cl.Close()
	}()

	poolAddress := common.HexToAddress("0x11b815efB8f581194ae79006d24E0d814B7697F6")

	pool, err := contracts.NewUniswapV3Pool(poolAddress, cl)

	if err != nil {
		slog.Error("Could not create uniswapv3pool")
	}

	swapChan := make(chan *contracts.UniswapV3PoolSwap)
	sub, err := pool.WatchSwap(&bind.WatchOpts{}, swapChan, nil, nil)
	if err != nil {
		slog.Error("Failed to subscribe to swap events", "error", err)
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			slog.Error("Subscription error", "error", err)
			return
		case log := <-swapChan:
			slog.Info("New log received", "log", log)

			// Process the log as needed
		}
	}

}
