package blockchain

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var (
	chains     map[string]*Network
	chainsOnce sync.Once
)

func init() {
	// Load .env before any package-level variables are initialized
	fmt.Println("Initializing blockchain package")
}

type Network struct {
	Network   string
	WsUrl     string
	HttpUrl   string
	ChainName string
}

func getChains() map[string]*Network {
	chainsOnce.Do(func() {
		// Load .env file
		godotenv.Load(".env")

		infuraAPIKey := os.Getenv("INFURA_API_KEY")

		chains = map[string]*Network{
			"BscMainnet": {
				Network:   "mainnet",
				WsUrl:     "wss://bsc-rpc.publicnode.com",
				HttpUrl:   "https://bsc-rpc.publicnode.com",
				ChainName: "BSC",
			},
			"BscMainnetInfura": {
				Network:   "mainnet",
				WsUrl:     "wss://bsc-mainnet.infura.io/ws/v3/" + infuraAPIKey,
				HttpUrl:   "https://bsc-rpc.publicnode.com",
				ChainName: "BSC",
			},
			"EthMainnet": {
				Network:   "mainnet",
				WsUrl:     "wss://mainnet.infura.io/ws/v3/" + infuraAPIKey,
				HttpUrl:   "https://ethereum.publicnode.com",
				ChainName: "Ethereum",
			},
		}
	})
	return chains
}

// Public getter for chains
func GetChains() map[string]*Network {
	return getChains()
}

var ActiveChain *Network

func Connect(network *Network) *ethclient.Client {
	if network != nil {
		ActiveChain = network
	} else {
		// Use default chain if none provided
		ActiveChain = getChains()["BscMainnet"]
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
