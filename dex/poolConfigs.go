package dex

import "math/big"

type Price struct {
	Pool            string
	Symbol          string
	Price           float64
	Liquidity       *big.Int
	LiquidityStatus string
}

type PoolConfig struct {
	token0         string
	token1         string
	token0Decimals int
	token1Decimals int
	Address        string
	TestAddress    string
}

var BnbUniswapv3SymbolToPool = map[string]*PoolConfig{
	"USDT/WBNB": {
		token0:         "USDT",
		token1:         "WBNB",
		token0Decimals: 6,
		token1Decimals: 18,
		Address:        "0x47a90A2d92A8367A91EfA1906bFc8c1E05bf10c4",
	},
}

var BnbPancakev3SymbolToPool = map[string]*PoolConfig{
	"USDT/WBNB": {
		token0:         "USDT",
		token1:         "WBNB",
		token0Decimals: 6,
		token1Decimals: 18,
		Address:        "0x172fcD41E0913e95784454622d1c3724f546f849",
	},
}

var BnbPancakeTestnetSymbolToPool = map[string]*PoolConfig{
	"WBNB/USDC": {
		token0:         "WBNB",
		token1:         "USDC",
		token0Decimals: 18,
		token1Decimals: 6,
		TestAddress:    "0x172fcD41E0913e95784454622d1c3724f546f849",
	},
	"BUSD/WBNB": {
		token0:         "BUSD",
		token1:         "WBNB",
		token0Decimals: 6,
		token1Decimals: 18,
		TestAddress:    "0x58C6Fc654b3deE6839b65136f61cB9120d96BCc6",
	},
	"USDT/WBNB": {
		token0:         "BUSD",
		token1:         "WBNB",
		token0Decimals: 6,
		token1Decimals: 18,
		TestAddress:    "0x5F52Ad4bD4f519AE79999400ad8B83A3D002fD92",
	},
}

var Uniswapv3SymbolToPool = map[string]*PoolConfig{
	"ETH/USDT": {
		token0:         "ETH",
		token1:         "USDT",
		token0Decimals: 18,
		token1Decimals: 6,
		Address:        "0x11b815efB8f581194ae79006d24E0d814B7697F6",
	},
	"WETH/USDC": {
		token0:         "WETH",
		token1:         "USDC",
		token0Decimals: 18,
		token1Decimals: 6,
		TestAddress:    "0x3289680dD4d6C10bb19b899729cda5eEF58AEfF1",
		Address:        "0x4e68Ccd3E89f51C3074ca5072bbAC773960dFa36",
	},
	"WETH/USDT": {
		token0:         "WETH",
		token1:         "USDT",
		token0Decimals: 18,
		token1Decimals: 6,
		Address:        "0x4e68Ccd3E89f51C3074ca5072bbAC773960dFa36",
	},
	"WBTC/WETH": {
		Address:        "0x4585FE77225b41b697C938B018E2Ac67Ac5a20c0",
		token0:         "WBTC",
		token1:         "WETH",
		token0Decimals: 8,
		token1Decimals: 18,
	},
}
var Pancakeswapv3SymbolToPool = map[string]*PoolConfig{
	"WETH/USDT": {
		token0:         "WETH",
		token1:         "USDT",
		token0Decimals: 18,
		token1Decimals: 6,
		Address:        "0x6CA298D2983aB03Aa1dA7679389D955A4eFEE15C",
	},
}
