package main

import (
	"fmt"
	"log"

	"merkle-generator/merkle"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	fmt.Println("=== Merkle Tree Library Example ===")

	// Example 1: Create leaves from raw data
	fmt.Println("\n1. Creating Merkle tree from raw data:")
	leaves := []common.Hash{
		merkle.HashData([]byte("alice")),
		merkle.HashData([]byte("bob")),
		merkle.HashData([]byte("charlie")),
		merkle.HashData([]byte("dave")),
	}

	fmt.Printf("Leaves:\n")
	for i, leaf := range leaves {
		fmt.Printf("  %d: %s\n", i, leaf.Hex())
	}

	// Create tree
	tree, err := merkle.NewMerkleTree(leaves)
	if err != nil {
		log.Fatal(err)
	}

	// Generate root
	root := tree.GenerateRoot()
	fmt.Printf("\nMerkle Root: %s\n", root.Hex())

	// Example 2: Generate and verify proof
	fmt.Println("\n2. Generating and verifying proof for 'alice':")
	target := leaves[0] // alice
	proof, err := tree.GenerateProof(target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Target (alice): %s\n", target.Hex())
	fmt.Printf("Proof elements:\n")
	for i, element := range proof {
		fmt.Printf("  %d: %s\n", i, element.Hex())
	}

	// Verify proof
	isValid := merkle.VerifyProof(proof, root, target)
	fmt.Printf("Proof is valid: %t\n", isValid)

	// Example 3: Test with invalid proof
	fmt.Println("\n3. Testing with invalid target:")
	invalidTarget := merkle.HashData([]byte("eve"))
	isValidInvalid := merkle.VerifyProof(proof, root, invalidTarget)
	fmt.Printf("Invalid proof result: %t\n", isValidInvalid)

	// Example 4: Working with hex data
	fmt.Println("\n4. Working with hex-encoded data:")
	hexLeaves := []common.Hash{
		common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"),
	}

	hexTree, err := merkle.NewMerkleTree(hexLeaves)
	if err != nil {
		log.Fatal(err)
	}

	hexRoot := hexTree.GenerateRoot()
	fmt.Printf("Hex tree root: %s\n", hexRoot.Hex())

	hexProof, err := hexTree.GenerateProof(hexLeaves[0])
	if err != nil {
		log.Fatal(err)
	}

	hexValid := merkle.VerifyProof(hexProof, hexRoot, hexLeaves[0])
	fmt.Printf("Hex proof valid: %t\n", hexValid)

	fmt.Println("\n=== Example completed successfully! ===")
}
