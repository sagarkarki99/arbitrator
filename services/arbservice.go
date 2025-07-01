package services

import (
	"log/slog"
	"math"

	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/dex"
)

var Amount = 15.0

type ArbService interface {
	Start()
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

func (a *ArbServiceImpl) Start() {
	price1, err := a.dex1.GetPrice("USDT/WBNB")
	if err != nil {
		slog.Error("Failed to get price from UNISWAP", "error", err)
	}
	price2, err := a.dex2.GetPrice("USDT/WBNB")
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
		case asset1Price := <-asset1:
			slog.Info("Received price from dex1", "price", asset1Price.Price)
			lastPrice1 = asset1Price.Price
			if lastPrice2 != 0 && a.IsProfit(lastPrice1, lastPrice2) {
				PerformArbitrageTransaction(lastPrice1, lastPrice2)
				// Perform arbitrage transaction
			}
		case asset2Price := <-asset2:
			slog.Info("Received price from dex2", "price", asset2Price.Price)
			lastPrice2 = asset2Price.Price
			if lastPrice1 != 0 && a.IsProfit(lastPrice2, lastPrice1) {
				// Perform arbitrage transaction
			}
		}
	}
}

func PerformArbitrageTransaction(lastPrice1, lastPrice2 float64) {
	if lastPrice1 < lastPrice2 {
		// buy in dex1 and sell in dex2
	} else {
		// buy in dex2 and sell in dex1
	}
}

func (a *ArbServiceImpl) IsProfit(price1, price2 float64) bool {
	diff := math.Abs(price1 - price2)
	gasFee, _ := blockchain.GetGasPrice()

	cost1 := (a.dex1.GetPoolFee() * price1) + float64(gasFee)
	cost2 := (a.dex2.GetPoolFee() * price2) + float64(gasFee)
	profit := diff - (cost1 + cost2)
	if profit <= 1 {
		// slog.Info("No profit opportunity found", "profit", profit, "price1", price1, "price2", price2)
		return false
	}
	return true
}
