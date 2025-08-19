// Package util provides reusable utilities for Ethereum interaction
package util

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient wraps ethclient.Client with additional utilities
type EthClient struct {
	*ethclient.Client
	ChainID *big.Int
}

// NewEthClient creates a new EthClient with chain ID cached
func NewEthClient(rpcURL string) (*EthClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Get and cache chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return &EthClient{
		Client:  client,
		ChainID: chainID,
	}, nil
}

// AccountInfo contains account-related information
type AccountInfo struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
	Nonce      uint64
}

// GetAccountInfo derives account information from a private key
func (ec *EthClient) GetAccountInfo(privateKeyHex string) (*AccountInfo, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(removeHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Get the public key and derive the address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get nonce
	nonce, err := ec.PendingNonceAt(context.Background(), address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	return &AccountInfo{
		Address:    address,
		PrivateKey: privateKey,
		Nonce:      nonce,
	}, nil
}

// TransactionParams contains parameters for creating transactions
type TransactionParams struct {
	To       common.Address
	Data     []byte
	Value    *big.Int
	GasLimit uint64
	GasPrice *big.Int
}

// EstimateAndCreateTransaction estimates gas and creates a transaction
func (ec *EthClient) EstimateAndCreateTransaction(account *AccountInfo, params TransactionParams) (*types.Transaction, error) {
	// Estimate gas if not provided
	if params.GasLimit == 0 {
		gasLimit, err := ec.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     account.Address,
			To:       &params.To,
			Data:     params.Data,
			Value:    params.Value,
			GasPrice: params.GasPrice,
		})
		if err != nil {
			// Use default gas limit if estimation fails
			params.GasLimit = 300000
		} else {
			params.GasLimit = gasLimit
		}
	}

	// Get gas price if not provided
	if params.GasPrice == nil {
		gasPrice, err := ec.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}
		params.GasPrice = gasPrice
	}

	// Set default value if not provided
	if params.Value == nil {
		params.Value = big.NewInt(0)
	}

	// Create transaction
	tx := types.NewTransaction(account.Nonce, params.To, params.Value, params.GasLimit, params.GasPrice, params.Data)

	return tx, nil
}

// SignAndSendTransaction signs a transaction and sends it to the network
func (ec *EthClient) SignAndSendTransaction(tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ec.ChainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = ec.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}

// removeHexPrefix removes the 0x prefix from hex strings
func removeHexPrefix(hex string) string {
	if len(hex) >= 2 && hex[:2] == "0x" {
		return hex[2:]
	}
	return hex
}
