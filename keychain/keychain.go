package keychain

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sagarkarki99/arbitrator/blockchain"
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
