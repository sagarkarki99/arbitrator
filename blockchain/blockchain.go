package blockchain

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Blockchain interface {
	Connect()
}

func Connect() *ethclient.Client {
	apiKey := os.Getenv("INFURA_API_KEY")
	wsUrl := fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s", apiKey)
	cl, err := ethclient.DialContext(context.Background(), wsUrl)
	if err != nil {
		panic(err)
	}
	slog.Info("Connected to Ethereum client", "url", wsUrl)
	return cl
}
