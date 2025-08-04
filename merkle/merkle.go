package merkle

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// MerkleTree represents a Merkle tree structure
type MerkleTree struct {
	leaves []common.Hash
}

// NewMerkleTree creates a new Merkle tree from the given leaves
func NewMerkleTree(leaves []common.Hash) (*MerkleTree, error) {
	if len(leaves) == 0 {
		return nil, errors.New("array cannot be empty")
	}

	// Create a copy of the leaves to avoid modifying the original slice
	leafCopy := make([]common.Hash, len(leaves))
	copy(leafCopy, leaves)

	return &MerkleTree{
		leaves: leafCopy,
	}, nil
}

// GenerateRoot generates the Merkle root from the leaves
func (mt *MerkleTree) GenerateRoot() common.Hash {
	if len(mt.leaves) == 1 {
		return mt.leaves[0]
	}

	currentLevel := make([]common.Hash, len(mt.leaves))
	copy(currentLevel, mt.leaves)

	// Build the Merkle Tree level by level
	for len(currentLevel) > 1 {
		nextLevel := make([]common.Hash, (len(currentLevel)+1)/2)

		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				// Pair exists, hash them together
				nextLevel[i/2] = hashPair(currentLevel[i], currentLevel[i+1])
			} else {
				// Odd leaf out, promote it to next level
				nextLevel[i/2] = currentLevel[i]
			}
		}
		currentLevel = nextLevel
	}

	return currentLevel[0]
}

// GenerateProof generates a Merkle proof for the target leaf
func (mt *MerkleTree) GenerateProof(target common.Hash) ([]common.Hash, error) {
	// Find the target index
	targetIndex := -1
	for i, leaf := range mt.leaves {
		if leaf == target {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return nil, errors.New("target leaf not found")
	}

	var proof []common.Hash
	currentLevel := make([]common.Hash, len(mt.leaves))
	copy(currentLevel, mt.leaves)
	currentTargetIndex := targetIndex

	// Build tree and collect proof
	for len(currentLevel) > 1 {
		nextLevel := make([]common.Hash, (len(currentLevel)+1)/2)
		nextTargetIndex := 0

		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				// Pair exists
				nextLevel[i/2] = hashPair(currentLevel[i], currentLevel[i+1])

				// If target is in this pair, add sibling to proof
				if i == (currentTargetIndex/2)*2 {
					if currentTargetIndex%2 == 0 {
						// Target is left child (even index), add right sibling
						proof = append(proof, currentLevel[i+1])
					} else {
						// Target is right child (odd index), add left sibling
						proof = append(proof, currentLevel[i])
					}
					nextTargetIndex = i / 2
				}
			} else {
				// Odd leaf
				nextLevel[i/2] = currentLevel[i]
				if i == (currentTargetIndex/2)*2 {
					nextTargetIndex = i / 2
				}
			}
		}
		currentLevel = nextLevel
		currentTargetIndex = nextTargetIndex
	}

	return proof, nil
}

// VerifyProof verifies if a leaf is in the Merkle Tree using the provided proof
func VerifyProof(proof []common.Hash, root common.Hash, target common.Hash) bool {
	computedHash := target

	for _, proofElement := range proof {
		computedHash = hashPair(computedHash, proofElement)
	}

	return computedHash == root
}

// hashPair hashes two nodes with consistent ordering (same as Solidity implementation)
func hashPair(a, b common.Hash) common.Hash {
	// Ensure a < b to maintain consistent ordering
	if bytes.Compare(a[:], b[:]) < 0 {
		return crypto.Keccak256Hash(append(a[:], b[:]...))
	}
	return crypto.Keccak256Hash(append(b[:], a[:]...))
}

// Helper function to convert hex string to common.Hash
func HexToHash(hex string) (common.Hash, error) {
	if !common.IsHexAddress(hex) && len(hex) != 66 { // 0x + 64 chars
		return common.Hash{}, fmt.Errorf("invalid hex string: %s", hex)
	}
	return common.HexToHash(hex), nil
}

// Helper function to create a hash from arbitrary data
func HashData(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}

// Helper function to create leaf hash like Solidity: keccak256(abi.encodePacked(address, amount))
func HashAddressAmount(address common.Address, amount *big.Int) common.Hash {
	// Convert amount to 32-byte array (like Solidity uint256)
	amountBytes := make([]byte, 32)
	amount.FillBytes(amountBytes) // This ensures exactly 32 bytes, big-endian

	// Concatenate address (20 bytes) + amount (32 bytes) = 52 bytes total
	data := append(address.Bytes(), amountBytes...)
	return crypto.Keccak256Hash(data)
}
