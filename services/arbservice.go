package services

import (
	"fmt"
	"log/slog"
	"math"
	"sync"

	"github.com/sagarkarki99/arbitrator/dex"
)

var DefaultOrderConfig = OrderConfig{
	AmountSize:      22.0,
	ProfitThreshold: 0.04,
	Slippage:        0.001,
	TotalGasCost:    0.0001,
	ActiveSymbol:    "CAKE/USDT",
}

// var DefaultOrderConfig = OrderConfig{
// 	AmountSize:      0.05,
// 	ProfitThreshold: 0.0001,
// 	Slippage:        0.001,
// 	TotalGasCost:    0.000000016,
// 	ActiveSymbol:    "USDT/WBNB",
// }

type OrderConfig struct {
	AmountSize      float64
	ProfitThreshold float64
	Slippage        float64
	TotalGasCost    float64
	ActiveSymbol    string
}

type ArbService interface {
	// Symbol is the trading pair symbol, e.g., "WETH/USDT"
	Start()

	SetConfig(newOrder OrderConfig)
}

type ArbServiceImpl struct {
	dex1 dex.Dex
	dex2 dex.Dex

	ConfigMutex *sync.RWMutex
	orderConfig OrderConfig
}

func NewArbService(dex1, dex2 dex.Dex, config OrderConfig) ArbService {
	return &ArbServiceImpl{
		dex1:        dex1,
		dex2:        dex2,
		ConfigMutex: &sync.RWMutex{},
		orderConfig: config,
	}
}

func (a *ArbServiceImpl) SetConfig(newOrder OrderConfig) {
	if newOrder.AmountSize <= 0 || newOrder.ProfitThreshold <= 0 {
		slog.Error("Invalid configuration values", "amountSize", newOrder.AmountSize, "profitThreshold", newOrder.ProfitThreshold)
		return
	}
	a.ConfigMutex.Lock()
	defer a.ConfigMutex.Unlock()
	a.orderConfig = newOrder

}

func (a *ArbServiceImpl) Start() {
	a.ConfigMutex.RLock()
	symbol := a.orderConfig.ActiveSymbol
	a.ConfigMutex.RUnlock()
	price1, err := a.dex1.GetPrice(symbol)
	if err != nil {
		slog.Error("Failed to get price from UNISWAP", "error", err)
	}
	price2, err := a.dex2.GetPrice(symbol)
	if err != nil {
		slog.Error("Failed to get price from PANCAKE", "error", err)
	}

	// TODO: run this function in separate go routine
	a.LookOpportunity(price1, price2)

}

func (a *ArbServiceImpl) LookOpportunity(asset1, asset2 <-chan *dex.Price) {
	lastPrice1 := 0.0
	lastPrice2 := 0.0
	for {
		select {
		case asset1Price, isOpen := <-asset1:
			if !isOpen {
				return
			}
			slog.Info(fmt.Sprintf("------- %s", asset1Price.Pool), "price", asset1Price.Price)
			lastPrice1 = asset1Price.Price
			if lastPrice2 != 0 && a.IsSpreadProfitable(lastPrice1, lastPrice2) && a.IsProfit(lastPrice1, lastPrice2) {
				a.performArbitrageTransaction(lastPrice1, lastPrice2, asset1Price.Symbol)
				// Perform arbitrage transaction
			}
		case asset2Price, isOpen := <-asset2:
			if !isOpen {
				return
			}
			slog.Info(fmt.Sprintf("------- %s", asset2Price.Pool), "price", asset2Price.Price)
			lastPrice2 = asset2Price.Price
			if lastPrice1 != 0 && a.IsSpreadProfitable(lastPrice1, lastPrice2) && a.IsProfit(lastPrice2, lastPrice1) {
				a.performArbitrageTransaction(lastPrice1, lastPrice2, asset2Price.Symbol)

			}
		}
	}
}

func (a *ArbServiceImpl) performArbitrageTransaction(lastPrice1, lastPrice2 float64, symbol string) {
	if math.Min(lastPrice1, lastPrice2) == lastPrice1 {
		// buy in dex1 and sell in dex2
		// a.dex1.Buy(amountSize, symbol)
		// a.dex2.Sell(amountSize, symbol)
		slog.Info("--------------------")
		slog.Info("Buy DEX1 / SELL DEX 2")
		slog.Info("--------------------")

	} else {
		slog.Info("--------------------")
		slog.Info("Buy DEX2 / SELL DEX1")
		slog.Info("--------------------")
		// a.dex2.Buy(amountSize, symbol)
		// a.dex1.Sell(amountSize, symbol)
		// buy in dex2 and sell in dex1
	}
}

func (a *ArbServiceImpl) IsSpreadProfitable(price1, price2 float64) bool {

	feeRatio := (a.dex1.GetPoolFee() + a.dex2.GetPoolFee()) + a.orderConfig.Slippage
	buyPrice := math.Min(price1, price2)
	sellPrice := math.Max(price1, price2)
	spreadRatio := ((sellPrice - buyPrice) / buyPrice)
	if spreadRatio > feeRatio {
		slog.Info("---Profitable ---- LET's Go0o0o0o0o0o0o-----")
	} else {
		slog.Info("---Not profitable---")
	}

	return spreadRatio > feeRatio

}

func (a *ArbServiceImpl) IsProfit(price1, price2 float64) bool {

	a.ConfigMutex.RLock()
	// Assumes AmountSize is in the quote currency (e.g., USDC).
	amountSize := a.orderConfig.AmountSize
	// Assumes ProfitThreshold is also in the quote currency (e.g., USDC).
	profitThreshold := a.orderConfig.ProfitThreshold
	a.ConfigMutex.RUnlock()

	buyPrice := 0.0
	sellPrice := 0.0
	buyFee := 0.0
	sellFee := 0.0
	if price1 > price2 {
		buyPrice = price2
		sellPrice = price1
		buyFee = a.dex2.GetPoolFee()
		sellFee = a.dex1.GetPoolFee()
	} else {
		buyPrice = price1
		sellPrice = price2
		sellFee = a.dex2.GetPoolFee()
		buyFee = a.dex1.GetPoolFee()
	}
	// -------BUYING (e.g., USDC -> WETH) ---------//
	// The fee is taken from the input asset (USDC) BEFORE the swap.
	netUsdcToSwap := amountSize * (1 - buyFee)
	wethReceived := netUsdcToSwap / buyPrice

	// -------SELLING (e.g., WETH -> USDC) ---------//
	// The fee is taken from the input asset (WETH) BEFORE the swap.
	netWethToSwap := wethReceived * (1 - sellFee)
	finalUsdcAmount := netWethToSwap * sellPrice

	// -------PROFIT CALCULATION (in USDC) ---------//
	// For this calculation to be accurate, `TotalGasCost` MUST be converted to,
	// and the `profitThreshold` MUST be in the same currency.
	// Since `TotalGasCost` is in USDC, `profitThreshold` should also be in USDC.
	// The `profitThreshold` is the amount of USDC you want to make.
	// The `Profit` calculation is `finalUsdcAmount - amountSize - TotalGasCost`.
	// This means `Profit` is the amount of USDC you make after all fees and gas.
	// If `Profit` is greater than or equal to `profitThreshold`, it's profitable.

	Profit := finalUsdcAmount - amountSize - a.orderConfig.TotalGasCost

	fmt.Println("----------------------------------------------------")
	slog.Info("Profit calculation (in USDC)",
		"buyPrice", buyPrice,
		"sellPrice", sellPrice,
		"amountSize_USDC", amountSize,
		"netUsdcToSwap", netUsdcToSwap,
		"wethReceived", wethReceived,
		"netWethToSwap", netWethToSwap,
		"finalUsdcAmount", finalUsdcAmount,
		"buyFee", buyFee,
		"sellFee", sellFee,
		"Profit", Profit,
		"profitThreshold", a.orderConfig.ProfitThreshold)
	fmt.Println("----------------------------------------------------")
	return Profit >= profitThreshold
}

// eg:
//  WETH/USDT
//  price1 = 1000
//  price2 = 1050
//  size = 500
//  cost1 = 0.003 * 500 + 0.0002 (gas fee)
// cost2= 0.003 * 500 + 0.0002 (gas fee)
// diff = 10
// size = 1000
// profit = size * diff - (cost1 + cost2)

// 1. make the transaction properties ready prehand:
//    - recipent address (my bot address)
//    - max gas limit (default 21000)
//    - gas price (default 12 gwei)
// 2. sign the transaction with private key
// 3. ready to send the transaction
// 4. Once there is a profit opportunity, send the transaction to the network
// 5. wait for the transaction to be mined
// 6. check the transaction receipt for success or failure
// 7. if success, repeat 1. if failed: keep same tranasction for next opportunity.
