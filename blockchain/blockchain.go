package blockchain

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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

	topic := []common.Hash{common.HexToHash("0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67")}
	poolAddress := common.HexToAddress("0x11b815efB8f581194ae79006d24E0d814B7697F6")
	query := ethereum.FilterQuery{
		Topics:    [][]common.Hash{topic},
		Addresses: []common.Address{poolAddress},
	}

	logChan := make(chan types.Log)
	sub, err := cl.SubscribeFilterLogs(context.Background(), query, logChan)
	if err != nil {
		slog.Error("Failed to subscribe to filter logs", "error", err)
	}
	slog.Info("Subscribed to filter logs", "address", poolAddress)
	defer func() {
		sub.Unsubscribe()
	}()

	for {
		select {
		case err := <-sub.Err():
			slog.Error("Subscription error", "error", err)
			return
		case log := <-logChan:
			slog.Info("New log received", "log", log)

			// Process the log as needed
		}
	}
}
