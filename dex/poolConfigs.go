package dex

import (
	"fmt"
	"math/big"

	"github.com/sagarkarki99/arbitrator/blockchain"
)

type Price struct {
	Pool            string
	Symbol          string
	Price           float64
	Liquidity       *big.Int
	LiquidityStatus string
}

type PoolConfig struct {
	Token0         string
	Token1         string
	Token0Contract string
	Token1Contract string
	Token0Decimals int
	Token1Decimals int
	Address        string
}

// NetworkConfig holds pool configurations for mainnet and testnet
type NetworkConfig struct {
	Mainnet map[DexApp]map[string]*PoolConfig
	Testnet map[DexApp]map[string]*PoolConfig
}

// ChainConfigs organizes pool configurations by chain
var ChainConfigs = map[string]*NetworkConfig{
	"ethereum": {
		Mainnet: map[DexApp]map[string]*PoolConfig{
			Uniswap: {
				"ETH/USDT": {
					Token0:         "ETH",
					Token1:         "USDT",
					Token0Decimals: 18,
					Token1Decimals: 6,
					Address:        "0x11b815efB8f581194ae79006d24E0d814B7697F6",
				},
				"WETH/USDC": {
					Token0:         "WETH",
					Token1:         "USDC",
					Token0Decimals: 18,
					Token1Decimals: 6,
					Address:        "0x4e68Ccd3E89f51C3074ca5072bbAC773960dFa36",
				},
				"WETH/USDT": {
					Token0:         "WETH",
					Token1:         "USDT",
					Token0Decimals: 18,
					Token1Decimals: 6,
					Address:        "0x4e68Ccd3E89f51C3074ca5072bbAC773960dFa36",
				},
				"WBTC/WETH": {
					Token0:         "WBTC",
					Token1:         "WETH",
					Token0Decimals: 8,
					Token1Decimals: 18,
					Address:        "0x4585FE77225b41b697C938B018E2Ac67Ac5a20c0",
				},
			},
			Pancakeswap: {
				"WETH/USDT": {
					Token0:         "WETH",
					Token1:         "USDT",
					Token0Decimals: 18,
					Token1Decimals: 6,
					Address:        "0x6CA298D2983aB03Aa1dA7679389D955A4eFEE15C",
				},
			},
		},
		Testnet: map[DexApp]map[string]*PoolConfig{
			Uniswap: {

				"USDC/WETH": {
					Token0:         "USDC",
					Token1:         "WETH",
					Token0Contract: "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238",
					Token1Contract: "0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14",
					Token0Decimals: 6,
					Token1Decimals: 18,
					Address:        "0x3289680dD4d6C10bb19b899729cda5eEF58AEfF1",
				},
			},
		},
	},
	"BSC": {
		Mainnet: map[DexApp]map[string]*PoolConfig{
			Uniswap: {
				"USDT/WBNB": {
					Token0:         "USDT",
					Token1:         "WBNB",
					Token0Decimals: 6,
					Token1Decimals: 18,
					Address:        "0x47a90A2d92A8367A91EfA1906bFc8c1E05bf10c4",
				},
			},
			Pancakeswap: {
				"USDT/WBNB": {
					Token0:         "USDT",
					Token1:         "WBNB",
					Token0Decimals: 6,
					Token1Decimals: 18,
					Address:        "0x172fcD41E0913e95784454622d1c3724f546f849",
				},
			},
		},
		Testnet: map[DexApp]map[string]*PoolConfig{
			Uniswap: {
				"USDT/WBNB": {
					Token0:         "USDT",
					Token1:         "WBNB",
					Token0Decimals: 6,
					Token1Decimals: 18,
					Address:        "0x47a90A2d92A8367A91EfA1906bFc8c1E05bf10c4",
				},
			},
			Pancakeswap: {
				"USDT/WBNB": {
					Token0:         "USDT",
					Token1:         "WBNB",
					Token0Decimals: 6,
					Token1Decimals: 18,
					Address:        "0x172fcD41E0913e95784454622d1c3724f546f849",
				},
			},
		},
	},
}

// GetActiveMarkets returns the pool configuration for the given symbol and dex
// based on the currently active blockchain network
func GetActiveMarkets(symbol string, dex DexApp) (*PoolConfig, error) {

	// Get the chain configuration
	chainConfig, exists := ChainConfigs[blockchain.ActiveChain.ChainName]
	if !exists {
		return nil, fmt.Errorf("chain configuration not found for: %s", blockchain.ActiveChain.ChainName)
	}

	// Select the appropriate network map
	var networkMap map[DexApp]map[string]*PoolConfig
	if blockchain.ActiveChain.Network == blockchain.Mainnet {
		networkMap = chainConfig.Mainnet
	} else {
		networkMap = chainConfig.Testnet
	}

	// Get the dex configuration
	dexConfig, exists := networkMap[dex]
	if !exists {
		return nil, fmt.Errorf("dex configuration not found for %s on %s %s",
			dex, blockchain.ActiveChain.ChainName, blockchain.ActiveChain.Network)
	}

	// Get the pool configuration
	poolConfig, exists := dexConfig[symbol]
	if !exists {
		return nil, fmt.Errorf("pool configuration not found for symbol %s on %s %s %s",
			symbol, dex, blockchain.ActiveChain.ChainName, blockchain.ActiveChain.Network)
	}

	return poolConfig, nil
}
