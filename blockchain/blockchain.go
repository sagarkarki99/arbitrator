package blockchain

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Network struct {
	Network   string
	WsUrl     string
	HttpUrl   string
	ChainName string
}

var Chains = map[string]*Network{
	"BscMainnet": {
		Network:   "mainnet",
		WsUrl:     "wss://bsc-rpc.publicnode.com",
		HttpUrl:   "https://bsc-rpc.publicnode.com",
		ChainName: "BSC",
	},
}
var ActiveChain = Chains["BscMainnet"]

func Connect(network *Network) *ethclient.Client {
	if network != nil {
		ActiveChain = network
	}
	cl, err := ethclient.DialContext(context.Background(), ActiveChain.WsUrl)
	if err != nil {
		panic(err)
	}
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
