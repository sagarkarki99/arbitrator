package blockchain

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var (
	Mainnet = "mainnet"
	Testnet = "testnet"
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
	ChainID   int
}

func getChains() map[string]*Network {
	chainsOnce.Do(func() {
		// Load .env file
		godotenv.Load(".env")

		infuraAPIKey := os.Getenv("INFURA_API_KEY")

		chains = map[string]*Network{
			"BscMainnet": {
				Network:   Mainnet,
				WsUrl:     "wss://bsc-rpc.publicnode.com",
				HttpUrl:   "https://bsc-rpc.publicnode.com",
				ChainName: "BSC",
				ChainID:   56,
			},
			"BscTestnet": {
				Network:   Testnet,
				WsUrl:     "wss://bsc-testnet-rpc.publicnode.com",
				HttpUrl:   "https://bsc-testnet.bnbchain.org",
				ChainName: "BSC",
				ChainID:   97,
			},
			"BscTestnetInfura": {
				Network:   Testnet,
				WsUrl:     "wss://bsc-testnet.infura.io/ws/v3/" + infuraAPIKey,
				HttpUrl:   "https://bsc-testnet.infura.io/v3/" + infuraAPIKey,
				ChainName: "BSC",
				ChainID:   97,
			},
			"BscMainnetInfura": {
				Network:   Mainnet,
				WsUrl:     "wss://bsc-mainnet.infura.io/ws/v3/" + infuraAPIKey,
				HttpUrl:   "https://bsc-rpc.publicnode.com",
				ChainName: "BSC",
				ChainID:   56,
			},
			"EthMainnet": {
				Network:   Mainnet,
				WsUrl:     "wss://mainnet.infura.io/ws/v3/" + infuraAPIKey,
				HttpUrl:   "https://ethereum.publicnode.com",
				ChainName: "ethereum",
				ChainID:   1,
			},
			"EthSepolia": {
				Network:   Testnet,
				WsUrl:     "wss://sepolia.infura.io/ws/v3/" + infuraAPIKey,
				HttpUrl:   "https://sepolia.infura.io/v3/" + infuraAPIKey,
				ChainName: "ethereum",
				ChainID:   11155111,
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

	slog.Info("Connected to chain",
		slog.Group("chain",
			"network", ActiveChain.Network,
			"wsUrl", ActiveChain.WsUrl,
			"httpUrl", ActiveChain.HttpUrl,
			"chainName", ActiveChain.ChainName,
			"chainID", ActiveChain.ChainID,
		),
	)
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
