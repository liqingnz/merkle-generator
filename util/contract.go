// Package util provides reusable contract interaction utilities
package util

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"merkle-generator/merkle"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// TokenClaimerABI contains the complete ABI for TokenClaimer contract
const TokenClaimerABI = `[
	{
		"inputs": [
			{
				"internalType": "bytes32[]",
				"name": "_proof",
				"type": "bytes32[]"
			},
			{
				"internalType": "bytes32",
				"name": "_root",
				"type": "bytes32"
			},
			{
				"internalType": "address",
				"name": "_addr",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "_amount",
				"type": "uint256"
			}
		],
		"name": "verifyAddress",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address[]",
				"name": "_addresses",
				"type": "address[]"
			},
			{
				"internalType": "uint256[]",
				"name": "_amounts",
				"type": "uint256[]"
			}
		],
		"name": "generateMerkleRoot",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address[]",
				"name": "_addresses",
				"type": "address[]"
			},
			{
				"internalType": "uint256[]",
				"name": "_amounts",
				"type": "uint256[]"
			},
			{
				"internalType": "address",
				"name": "_target",
				"type": "address"
			}
		],
		"name": "generateProof",
		"outputs": [
			{
				"internalType": "bytes32[]",
				"name": "",
				"type": "bytes32[]"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "_amount",
				"type": "uint256"
			},
			{
				"internalType": "bytes32[]",
				"name": "_proof",
				"type": "bytes32[]"
			}
		],
		"name": "claim",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	}
]`

// TokenClaimerContract wraps contract interactions
type TokenClaimerContract struct {
	Address common.Address
	ABI     abi.ABI
	Client  *EthClient
}

// NewTokenClaimerContract creates a new TokenClaimer contract wrapper
func NewTokenClaimerContract(contractAddress string, client *EthClient) (*TokenClaimerContract, error) {
	// Parse contract ABI
	contractABI, err := abi.JSON(strings.NewReader(TokenClaimerABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	return &TokenClaimerContract{
		Address: common.HexToAddress(contractAddress),
		ABI:     contractABI,
		Client:  client,
	}, nil
}

// TestCase represents a test case for merkle tree operations
type TestCase struct {
	Name    string
	Address common.Address
	Amount  *big.Int
}

// GenerateMerkleRoot calls the contract's generateMerkleRoot function
func (tc *TokenClaimerContract) GenerateMerkleRoot(testCases []TestCase) (common.Hash, error) {
	// Prepare arrays for contract calls
	addresses := make([]common.Address, len(testCases))
	amounts := make([]*big.Int, len(testCases))
	for i, testCase := range testCases {
		addresses[i] = testCase.Address
		amounts[i] = testCase.Amount
	}

	// Pack the function call
	data, err := tc.ABI.Pack("generateMerkleRoot", addresses, amounts)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack generateMerkleRoot call: %w", err)
	}

	// Call the contract
	callMsg := ethereum.CallMsg{
		To:   &tc.Address,
		Data: data,
	}

	result, err := tc.Client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to call generateMerkleRoot: %w", err)
	}

	// Unpack the result
	var contractRoot common.Hash
	err = tc.ABI.UnpackIntoInterface(&contractRoot, "generateMerkleRoot", result)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to unpack generateMerkleRoot result: %w", err)
	}

	return contractRoot, nil
}

// GenerateProof calls the contract's generateProof function
func (tc *TokenClaimerContract) GenerateProof(testCases []TestCase, targetAddress common.Address) ([]common.Hash, error) {
	// Prepare arrays for contract calls
	addresses := make([]common.Address, len(testCases))
	amounts := make([]*big.Int, len(testCases))
	for i, testCase := range testCases {
		addresses[i] = testCase.Address
		amounts[i] = testCase.Amount
	}

	// Pack the function call
	data, err := tc.ABI.Pack("generateProof", addresses, amounts, targetAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to pack generateProof call: %w", err)
	}

	// Call the contract
	callMsg := ethereum.CallMsg{
		To:   &tc.Address,
		Data: data,
	}

	result, err := tc.Client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call generateProof: %w", err)
	}

	// Unpack the result
	var contractProof []common.Hash
	err = tc.ABI.UnpackIntoInterface(&contractProof, "generateProof", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack generateProof result: %w", err)
	}

	return contractProof, nil
}

// VerifyAddress calls the contract's verifyAddress function
func (tc *TokenClaimerContract) VerifyAddress(proof []common.Hash, root common.Hash, address common.Address, amount *big.Int) (bool, error) {
	// Pack the function call
	data, err := tc.ABI.Pack("verifyAddress", proof, root, address, amount)
	if err != nil {
		return false, fmt.Errorf("failed to pack verifyAddress call: %w", err)
	}

	// Call the contract
	callMsg := ethereum.CallMsg{
		To:   &tc.Address,
		Data: data,
	}

	result, err := tc.Client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return false, fmt.Errorf("failed to call verifyAddress: %w", err)
	}

	// Unpack the result
	var isValid bool
	err = tc.ABI.UnpackIntoInterface(&isValid, "verifyAddress", result)
	if err != nil {
		return false, fmt.Errorf("failed to unpack verifyAddress result: %w", err)
	}

	return isValid, nil
}

// PrepareClaimTransaction prepares a claim transaction
func (tc *TokenClaimerContract) PrepareClaimTransaction(to common.Address, amount *big.Int, proof []common.Hash) ([]byte, error) {
	// Pack the claim function call
	data, err := tc.ABI.Pack("claim", to, amount, proof)
	if err != nil {
		return nil, fmt.Errorf("failed to pack claim function call: %w", err)
	}

	return data, nil
}

// MerkleData contains all merkle tree related data
type MerkleData struct {
	Root   common.Hash
	Leaves []common.Hash
	Tree   *merkle.MerkleTree
}

// GenerateLocalMerkleData generates local merkle tree data from test cases
func GenerateLocalMerkleData(testCases []TestCase) (*MerkleData, error) {
	// Generate leaves from test cases
	leaves := make([]common.Hash, len(testCases))
	for i, testCase := range testCases {
		leaves[i] = merkle.HashAddressAmount(testCase.Address, testCase.Amount)
	}

	// Create Merkle tree
	tree, err := merkle.NewMerkleTree(leaves)
	if err != nil {
		return nil, fmt.Errorf("failed to create merkle tree: %w", err)
	}

	root := tree.GenerateRoot()

	return &MerkleData{
		Root:   root,
		Leaves: leaves,
		Tree:   tree,
	}, nil
}

// GenerateLocalProof generates a proof for a specific test case index
func (md *MerkleData) GenerateLocalProof(index int) ([]common.Hash, error) {
	if index < 0 || index >= len(md.Leaves) {
		return nil, fmt.Errorf("invalid index: %d", index)
	}

	proof, err := md.Tree.GenerateProof(md.Leaves[index])
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	return proof, nil
}

// FindTestCaseIndex finds the index of a test case by address
func FindTestCaseIndex(testCases []TestCase, address common.Address) int {
	for i, testCase := range testCases {
		if testCase.Address == address {
			return i
		}
	}
	return -1
}
