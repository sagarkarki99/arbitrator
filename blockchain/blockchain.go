package blockchain

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/ethclient"
)

var Testnet = "bsc-testnet"
var Mainnet = "bsc-mainnet"
var Network = Mainnet

type Blockchain interface {
	Connect()
}

func Connect() *ethclient.Client {
	// apiKey := os.Getenv("INFURA_API_KEY")
	// wsUrl := fmt.Sprintf("wss://%s.infura.io/ws/v3/%s", Network, apiKey)
	wsUrl := "wss://bsc-rpc.publicnode.com"
	// wsUrl := "wss://bsc-testnet.drpc.org"
	cl, err := ethclient.DialContext(context.Background(), wsUrl)
	if err != nil {
		panic(err)
	}
	slog.Info("Connected to BNB Chain client", "url", wsUrl)
	return cl
}

func GetGasPrice() (uint64, error) {
	return 12, nil
	// httpClient, _ := ethclient.Dial("https://ethereum.publicnode.com")
	// gasPrice, err := httpClient.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to suggest gas price: %w", err)
	// }
	// slog.Info("Suggested gas price", "gasPrice", gasPrice)
	// return gasPrice.Uint64(), nil
}
