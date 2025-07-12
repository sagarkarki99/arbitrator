package keychain

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/contracts"
)

var Accounts []string = []string{
	"0x42E45cC2929312cfE071eDF35F423434ED92E316",
	"0x267f5BFc9486b446046e97F9D3864793dA8eFb33",
}

// to is address that is spender, tokenContract is the ERC20 token contract address
func SendApproval(cl *ethclient.Client, to string, tokenContract string, kc Keychain) bool {

	tokenAddress := common.HexToAddress(tokenContract)

	// Check if token contract exists
	code, err := cl.CodeAt(context.Background(), tokenAddress, nil)
	if err != nil {
		slog.Error("Failed to check token contract", "error", err)
		return true
	}
	if len(code) == 0 {
		slog.Error("Token contract does not exist", "address", tokenContract)
		return true
	}

	erc20, err := contracts.NewERC20(tokenAddress, cl)
	if err != nil {
		slog.Error("Failed to create ERC20 contract", "error", err)
		return true
	}

	poolAddress := common.HexToAddress(to)
	myAddress := common.HexToAddress(Accounts[0])

	// Get nonce for transaction
	nonce, err := cl.PendingNonceAt(context.Background(), myAddress)
	if err != nil {
		slog.Error("Failed to get nonce", "error", err)
		return true
	}

	// Check account balance
	balance, err := cl.BalanceAt(context.Background(), myAddress, nil)
	if err != nil {
		slog.Error("Failed to get account balance", "error", err)
		return true
	}

	minBalance := big.NewInt(1e16) // 0.01 ETH minimum (1e18 wei = 1 ETH, so 1e16 wei = 0.01 ETH)
	if balance.Cmp(minBalance) < 0 {
		slog.Error("Insufficient ETH balance for gas fees",
			"balance", balance,
			"required", minBalance,
			"account", myAddress.Hex())
		return true
	}

	slog.Info("Account info",
		"address", myAddress.Hex(),
		"balance", balance,
		"nonce", nonce)

	auth := &bind.TransactOpts{
		Nonce:     big.NewInt(int64(nonce)),
		From:      myAddress,
		GasLimit:  100000,                      // Set gas limit for approval
		GasFeeCap: big.NewInt(100_000_000_000), // 100 Gwei
		GasTipCap: big.NewInt(4_000_000_000),   // 4 Gwei
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return kc.Sign(tx)
		},
	}

	slog.Info("Approving token transfer",
		"token", tokenContract,
		"spender", to,
		"amount", "1e18")

	trx, err := erc20.Approve(auth, poolAddress, big.NewInt(1e18))
	if err != nil {
		slog.Error("Failed to approve token transfer", "error", err)
		return true
	}

	fmt.Println("Transaction hash:", trx.Hash().Hex())
	return false
}
