package services

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/sagarkarki99/arbitrator/dex"
)

var Amount = 15.0

type ArbService interface {
	// Symbol is the trading pair symbol, e.g., "WETH/USDT"
	Start(symbol string)
}

type ArbServiceImpl struct {
	dex1 dex.Dex
	dex2 dex.Dex
}

func NewArbService(dex1, dex2 dex.Dex) ArbService {
	return &ArbServiceImpl{
		dex1: dex1,
		dex2: dex2,
	}
}

func (a *ArbServiceImpl) Start(symbol string) {
	price1, err := a.dex1.GetPrice(symbol)
	if err != nil {
		slog.Error("Failed to get price from UNISWAP", "error", err)
	}
	price2, err := a.dex2.GetPrice(symbol)
	if err != nil {
		slog.Error("Failed to get price from PANCAKE", "error", err)
	}

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
			slog.Info("Received price from dex1", "price", asset1Price.Price)
			lastPrice1 = asset1Price.Price
			if lastPrice2 != 0 && a.IsSpreadProfitable(lastPrice1, lastPrice2) && a.IsProfit(lastPrice1, lastPrice2) {
				a.performArbitrageTransaction(lastPrice1, lastPrice2)
				// Perform arbitrage transaction

			}
		case asset2Price, isOpen := <-asset2:
			if !isOpen {
				return
			}
			slog.Info("Received price from dex2", "price", asset2Price.Price)
			lastPrice2 = asset2Price.Price
			if lastPrice1 != 0 && a.IsSpreadProfitable(lastPrice1, lastPrice2) && a.IsProfit(lastPrice2, lastPrice1) {
				// Perform arbitrage transaction
			}
		}
	}
}

func (a *ArbServiceImpl) performArbitrageTransaction(lastPrice1, lastPrice2 float64) {
	if math.Min(lastPrice1, lastPrice2) == lastPrice1 {
		// buy in dex1 and sell in dex2
		a.dex1.Buy(amountSize)
		a.dex2.Sell(amountSize)

	} else {
		a.dex2.Buy(amountSize)
		a.dex1.Sell(amountSize)
		// buy in dex2 and sell in dex1
	}
}

func (a *ArbServiceImpl) IsSpreadProfitable(price1, price2 float64) bool {

	totalFeePercent := (a.dex1.GetPoolFee() + a.dex2.GetPoolFee()) + 0.1
	buyPrice := math.Min(price1, price2)
	sellPrice := math.Max(price1, price2)
	spreadPercent := ((sellPrice - buyPrice) / buyPrice) * 100
	return spreadPercent >= totalFeePercent

}

func (a *ArbServiceImpl) IsProfit(price1, price2 float64) bool {

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
	// -------BUYING ---------//
	netBuyingSize := amountSize - (buyFee * amountSize)
	tokenReceived := (1 / buyPrice) * netBuyingSize

	// -------SELLING ---------//
	netSellingSize := tokenReceived - (sellFee * tokenReceived)
	baseToken := netSellingSize * sellPrice

	// -------PROFIT CALCULATION ---------//

	Profit := baseToken - amountSize - totalGasCost

	fmt.Println("----------------------------------------------------")
	slog.Info("Profit calculation",
		"buyPrice", buyPrice,
		"sellPrice", sellPrice,
		"amountSize", amountSize,
		"buyFee", buyFee,
		"sellFee", sellFee,
		"Profit", Profit,
		"profitThreshold", profitThreshold)
	fmt.Println("----------------------------------------------------")
	return Profit >= profitThreshold
}

var totalGasCost = 0.0016
var amountSize = 20000.0
var profitThreshold = 2.00 // USDT
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
