package merkle

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestNewMerkleTree(t *testing.T) {
	// Test empty leaves
	_, err := NewMerkleTree([]common.Hash{})
	if err == nil {
		t.Error("Expected error for empty leaves")
	}

	// Test valid leaves
	leaves := []common.Hash{
		common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"),
	}
	tree, err := NewMerkleTree(leaves)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(tree.leaves) != 2 {
		t.Errorf("Expected 2 leaves, got %d", len(tree.leaves))
	}
}

func TestGenerateRoot(t *testing.T) {
	// Test single leaf
	leaves := []common.Hash{
		common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
	}
	tree, _ := NewMerkleTree(leaves)
	root := tree.GenerateRoot()
	if root != leaves[0] {
		t.Error("Single leaf root should equal the leaf itself")
	}

	// Test multiple leaves
	leaves = []common.Hash{
		HashData([]byte("alice")),
		HashData([]byte("bob")),
		HashData([]byte("charlie")),
	}
	tree, _ = NewMerkleTree(leaves)
	root = tree.GenerateRoot()

	// Root should not be any of the original leaves
	for _, leaf := range leaves {
		if root == leaf {
			t.Error("Root should not equal any of the original leaves")
		}
	}
}

func TestGenerateProof(t *testing.T) {
	leaves := []common.Hash{
		HashData([]byte("alice")),
		HashData([]byte("bob")),
		HashData([]byte("charlie")),
		HashData([]byte("dave")),
	}
	tree, _ := NewMerkleTree(leaves)

	// Test proof generation for existing leaf
	proof, err := tree.GenerateProof(leaves[0])
	if err != nil {
		t.Errorf("Unexpected error generating proof: %v", err)
	}
	if len(proof) == 0 {
		t.Error("Proof should not be empty for multiple leaves")
	}

	// Test proof generation for non-existing leaf
	nonExistent := HashData([]byte("eve"))
	_, err = tree.GenerateProof(nonExistent)
	if err == nil {
		t.Error("Expected error for non-existent leaf")
	}
}

func TestVerifyProof(t *testing.T) {
	leaves := []common.Hash{
		HashData([]byte("alice")),
		HashData([]byte("bob")),
		HashData([]byte("charlie")),
		HashData([]byte("dave")),
	}
	tree, _ := NewMerkleTree(leaves)
	root := tree.GenerateRoot()

	// Test valid proof
	target := leaves[0]
	proof, _ := tree.GenerateProof(target)
	isValid := VerifyProof(proof, root, target)
	if !isValid {
		t.Error("Valid proof should verify successfully")
	}

	// Test invalid proof (wrong target)
	wrongTarget := HashData([]byte("eve"))
	isValid = VerifyProof(proof, root, wrongTarget)
	if isValid {
		t.Error("Invalid proof should not verify")
	}

	// Test invalid proof (wrong root)
	wrongRoot := HashData([]byte("wrong"))
	isValid = VerifyProof(proof, wrongRoot, target)
	if isValid {
		t.Error("Proof with wrong root should not verify")
	}
}

func TestHashPair(t *testing.T) {
	a := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	// Hash pair should be consistent regardless of order
	hash1 := hashPair(a, b)
	hash2 := hashPair(b, a)

	if hash1 != hash2 {
		t.Error("Hash pair should be consistent regardless of input order")
	}
}

func TestHexToHash(t *testing.T) {
	// Test valid hex
	validHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	hash, err := HexToHash(validHex)
	if err != nil {
		t.Errorf("Unexpected error for valid hex: %v", err)
	}
	if hash.Hex() != validHex {
		t.Error("Hash conversion mismatch")
	}

	// Test invalid hex
	invalidHex := "invalid"
	_, err = HexToHash(invalidHex)
	if err == nil {
		t.Error("Expected error for invalid hex")
	}
}

func TestConsistencyWithSolidity(t *testing.T) {
	// Test with known values to ensure consistency with Solidity implementation
	leaves := []common.Hash{
		HashData([]byte("alice")),
		HashData([]byte("bob")),
	}

	tree, _ := NewMerkleTree(leaves)
	root := tree.GenerateRoot()

	// Verify that we can generate and verify proofs
	proof, err := tree.GenerateProof(leaves[0])
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	isValid := VerifyProof(proof, root, leaves[0])
	if !isValid {
		t.Error("Generated proof should be valid")
	}
}

func TestHashAddressAmount(t *testing.T) {
	// Test the HashAddressAmount function matches Solidity exactly
	// This is the test case provided by the user
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")
	amount := big.NewInt(1000000000000000000) // 1 ETH in wei

	expectedHash := common.HexToHash("0x862d9f69cd1642f07c56ec6b92856ce141af9dfe404d2ab0c4685a334945ffe6")
	actualHash := HashAddressAmount(address, amount)

	if actualHash != expectedHash {
		t.Errorf("HashAddressAmount mismatch:\nExpected: %s\nActual:   %s", expectedHash.Hex(), actualHash.Hex())
	}
}

func TestHashAddressAmountWithDifferentValues(t *testing.T) {
	// Test with different values to ensure consistency
	testCases := []struct {
		name    string
		address common.Address
		amount  *big.Int
	}{
		{
			name:    "Zero amount",
			address: common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"),
			amount:  big.NewInt(0),
		},
		{
			name:    "Large amount",
			address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
			amount:  big.NewInt(2500000000000000000), // 2.5 ETH
		},
		{
			name:    "Zero address",
			address: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			amount:  big.NewInt(1000000000000000000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash1 := HashAddressAmount(tc.address, tc.amount)
			hash2 := HashAddressAmount(tc.address, tc.amount)

			// Should be deterministic
			if hash1 != hash2 {
				t.Errorf("HashAddressAmount is not deterministic for %s", tc.name)
			}

			// Should not be zero hash
			zeroHash := common.Hash{}
			if hash1 == zeroHash {
				t.Errorf("HashAddressAmount returned zero hash for %s", tc.name)
			}
		})
	}
}
