package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/sagarkarki99/arbitrator/blockchain"
)

func main() {
	fmt.Println("Hello arbitrator")
	godotenv.Load(
		".env",
	)

	blockchain.Connect()
}
