package services

import (
	"sync"
	"testing"

	"github.com/sagarkarki99/arbitrator/dex"
)

type MockDex1 struct {
}

func (m1 MockDex1) GetPrice(symbol string) (<-chan *dex.Price, error) {
	return nil, nil
}

func (m1 MockDex1) GetPoolFee() float64 {
	return 0.003
}

func (m1 MockDex1) Buy(amount float64, symbol string) (string, error) {
	return "", nil
}
func (m1 MockDex1) Sell(amount float64, symbol string) (string, error) {
	return "", nil
}

type MockDex2 struct {
}

func (m1 MockDex2) GetPrice(symbol string) (<-chan *dex.Price, error) {
	return nil, nil
}

func (m1 MockDex2) GetPoolFee() float64 {
	return 0.003
}

func (m1 MockDex2) Buy(amount float64, symbol string) (string, error) {
	return "", nil
}
func (m1 MockDex2) Sell(amount float64, symbol string) (string, error) {
	return "", nil
}

func TestIsProfit_AsTrue(t *testing.T) {
	// Arrange
	mockProfitThreshold := 5.0
	mockAmountSize := 100.0
	mockDex1 := MockDex1{}
	mockDex2 := MockDex2{}
	arbService := &ArbServiceImpl{
		dex1:        mockDex1,
		dex2:        mockDex2,
		ConfigMutex: &sync.RWMutex{},
	}
	// profit = (price1 - price2) - (price1 * dex1.GetPoolFee() + price2 * dex2.GetPoolFee()) - TotalGasCost
	// profit = 50 - (100 * 0.003)
	arbService.SetConfig(mockAmountSize, mockProfitThreshold)

	mockPrice1 := 100.0
	mockPrice2 := 90.0

	// Act
	isProfit := arbService.IsProfit(mockPrice1, mockPrice2)

	// Assert

	if isProfit == false {
		t.Errorf("Expected IsProfit to return true, but got false")
	}
	t.Logf("IsProfit returned true as expected for prices: %f and %f", mockPrice1, mockPrice2)

}

func TestIsProfit_AsFalse(t *testing.T) {
	// Arrange
	mockProfitThreshold := 5.0
	mockAmountSize := 100.0
	mockDex1 := MockDex1{}
	mockDex2 := MockDex2{}
	mockPrice1 := 100.0
	mockPrice2 := 98.0

	arbService := &ArbServiceImpl{
		dex1:        mockDex1,
		dex2:        mockDex2,
		ConfigMutex: &sync.RWMutex{},
	}

	arbService.SetConfig(mockAmountSize, mockProfitThreshold)

	// Act
	isProfit := arbService.IsProfit(mockPrice1, mockPrice2)

	// Assert

	if isProfit == true {
		t.Errorf("Expected IsProfit as %t, but got %t", false, isProfit)
	}
	t.Logf("IsProfit returned true as expected for prices: %f and %f", mockPrice1, mockPrice2)

}

func TestIsSpreadProfitable_AsTrue(t *testing.T) {
	// Arrange

	arbService := &ArbServiceImpl{
		dex1:        MockDex1{},
		dex2:        MockDex2{},
		ConfigMutex: &sync.RWMutex{},
	}
	mockPrice1 := 100.0
	mockPrice2 := 99.3
	mockAmountSize := 100.0
	mockProfitThreshold := 90.0
	arbService.SetConfig(mockAmountSize, mockProfitThreshold) // Set amount size and profit threshold

	// Act
	// total cost should be smaller than price difference
	isProfitable := arbService.IsSpreadProfitable(mockPrice1, mockPrice2)

	// Assert
	if isProfitable == false {
		t.Errorf("Expected IsSpreadProfitable as %t but received %t", true, isProfitable)
	}

}
func TestIsSpreadProfitable_AsFalse(t *testing.T) {
	// Arrange

	arbService := &ArbServiceImpl{
		dex1:        MockDex1{},
		dex2:        MockDex2{},
		ConfigMutex: &sync.RWMutex{},
	}
	mockAmountSize := 100.0
	mockProfitThreshold := 90.0
	mockPrice1 := 100.0
	mockPrice2 := 99.5
	arbService.SetConfig(mockAmountSize, mockProfitThreshold) // Set amount size and profit threshold

	// Act
	// total cost should be smaller than price difference
	isProfitable := arbService.IsSpreadProfitable(mockPrice1, mockPrice2)

	// Assert
	if isProfitable == true {
		t.Errorf("Expected IsSpreadProfitable as %t but received %t", true, isProfitable)
	}

}
