package dex

type Dex interface {
	GetPrice(symbol string) (float64, error)
	CreateTransaction(amount float64, symbol string, from string) (string, error)
}
