package keychain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sagarkarki99/arbitrator/blockchain"
	"github.com/sagarkarki99/arbitrator/constants"
	"github.com/sagarkarki99/arbitrator/contracts"
)

type Keychain interface {
	Sign(trx *types.Transaction) (*types.Transaction, error)
}

func NewKeychainImpl() Keychain {
	return &KeychainImpl{}
}

type KeychainImpl struct {
}

func (k *KeychainImpl) Sign(trx *types.Transaction) (*types.Transaction, error) {
	pk := os.Getenv("pk")
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, errors.New("failed to sign transaction")
	}
	pubkey := privateKey.Public()
	pubkeyECDSA, ok := pubkey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to cast public key to ECDSA")
	}
	slog.Info("From address:", "address", crypto.PubkeyToAddress(*pubkeyECDSA).Hex())

	signer := types.LatestSignerForChainID(big.NewInt(int64(blockchain.ActiveChain.ChainID)))
	signedTx, err := types.SignTx(trx, signer, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signedTx, nil
}

func GetBalance(address string, tokenContract string, cl *ethclient.Client) (*big.Int, error) {
	myAddress := common.HexToAddress(Accounts[0])
	contract := common.HexToAddress(tokenContract)

	erc20, err := contracts.NewERC20(contract, cl)
	if err != nil {
		slog.Error("Failed to create USDC contract", "error", err)
		return nil, fmt.Errorf("failed to create ERC20 contract: %w", err)
	}

	balance, err := erc20.BalanceOf(&bind.CallOpts{}, myAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func ConvertReadable(balance *big.Int, decimal int) *big.Float {
	balanceFloat := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)) // 10^decimal
	humanBalance := new(big.Float).Quo(balanceFloat, divisor)
	return humanBalance
}

func GetNativeBalance(address string, cl *ethclient.Client) (*big.Int, error) {

	myAddress := common.HexToAddress(address)
	ethBalance, err := cl.BalanceAt(context.Background(), myAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH balance: %w", err)
	}

	ethBalanceFloat := new(big.Float).SetInt(ethBalance)
	ethDivisor := new(big.Float).SetInt(big.NewInt(1e18)) // 18 decimals for ETH
	ethHuman := new(big.Float).Quo(ethBalanceFloat, ethDivisor)
	fmt.Printf("ETH Balance: %s ETH\n", ethHuman.String())
	return ethBalance, nil
}

func HasApproval(address string, cl *ethclient.Client) bool {
	myAddress := common.HexToAddress(Accounts[0])
	poolAddress := common.HexToAddress(address)
	erc20, _ := contracts.NewERC20(poolAddress, cl)
	allowance, err := erc20.Allowance(&bind.CallOpts{}, myAddress, poolAddress)
	if err != nil {
		slog.Error("Failed to get allowance", "error", err)
		return false
	}

	fmt.Printf("Pool allowance: %s USDC\n", allowance.String())

	if allowance.Cmp(big.NewInt(0)) == 0 {
		slog.Error("Pool has no allowance to spend your USDC! Run approval first.")
		return false
	}
	return true
}

func WrapNative(amount float64, wrapperContract string, cl *ethclient.Client, kc Keychain) {
	ethAmount := amount
	weiAmount := new(big.Float).SetFloat64(ethAmount)
	weiAmount.Mul(weiAmount, big.NewFloat(1e18))
	eth, _ := weiAmount.Int(new(big.Int))

	WrappedAddress := common.HexToAddress(wrapperContract)
	con, _ := contracts.NewERC20(WrappedAddress, cl)
	trx, err := con.Deposit(&bind.TransactOpts{
		From:      common.HexToAddress(Accounts[0]),
		Value:     eth, // ETH amount to wrap
		GasFeeCap: constants.GasFeeCap,
		GasTipCap: constants.GasTipCap,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return kc.Sign(tx)
		},
	})

	slog.Info("Deposit transaction sent", "hash", trx.Hash().Hex())

	if err != nil {
		slog.Error("Failed to wrap ETH to WETH", "error", err)
		return
	}
}
