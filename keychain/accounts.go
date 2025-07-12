package keychain

import (
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

func SendApproval(cl *ethclient.Client, to string, tokenContract string, kc Keychain) bool {
	addr := common.HexToAddress(to)
	erc20, err := contracts.NewERC20(addr, cl)
	if err != nil {
		slog.Error("Failed to create ERC20 contract", "error", err)
		return true
	}

	poolAddress := common.HexToAddress(tokenContract)
	auth := &bind.TransactOpts{
		Nonce:     big.NewInt(17),
		From:      common.HexToAddress(Accounts[0]),
		GasFeeCap: big.NewInt(10_000_000_000), // 10 Gwei
		GasTipCap: big.NewInt(10_000_000_000), // 10 Gwei
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return kc.Sign(tx)
		},
	}
	trx, err := erc20.Approve(auth, poolAddress, big.NewInt(1000000000000))
	if err != nil {
		slog.Error("Failed to approve token transfer", "error", err)
		return true

	}

	fmt.Println("Transaction hash:", trx.Hash().Hex())
	return false
}
