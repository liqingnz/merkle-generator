package main

import (
	"fmt"
	"log"
	"math/big"

	"merkle-generator/merkle"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	fmt.Println("=== TokenClaimer Contract Test Data ===")
	fmt.Println("This script generates test data compatible with your TokenClaimer.sol contract")
	fmt.Println()

	// Test data - addresses and amounts
	testData := []struct {
		address common.Address
		amount  *big.Int
		name    string
	}{
		{
			address: common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"),
			amount:  big.NewInt(1000000000000000000), // 1 ETH
			name:    "Alice",
		},
		{
			address: common.HexToAddress("0x8ba1f109551bD432803012645Hac136c772c3e3C"),
			amount:  big.NewInt(500000000000000000), // 0.5 ETH
			name:    "Bob",
		},
		{
			address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
			amount:  big.NewInt(2500000000000000000), // 2.5 ETH
			name:    "Charlie",
		},
		{
			address: common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
			amount:  big.NewInt(750000000000000000), // 0.75 ETH
			name:    "Dave",
		},
	}

	// Create leaves
	fmt.Println("=== STEP 1: Creating Leaves ===")
	leaves := make([]common.Hash, len(testData))
	for i, data := range testData {
		leaves[i] = merkle.HashAddressAmount(data.address, data.amount)
		fmt.Printf("%s:\n", data.name)
		fmt.Printf("  Address: %s\n", data.address.Hex())
		fmt.Printf("  Amount:  %s wei\n", data.amount.String())
		fmt.Printf("  Leaf:    %s\n", leaves[i].Hex())
		fmt.Println()
	}

	// Create Merkle tree
	fmt.Println("=== STEP 2: Creating Merkle Tree ===")
	tree, err := merkle.NewMerkleTree(leaves)
	if err != nil {
		log.Fatal(err)
	}

	root := tree.GenerateRoot()
	fmt.Printf("Merkle Root: %s\n", root.Hex())
	fmt.Println()

	// Generate proofs for each address
	fmt.Println("=== STEP 3: Generating Proofs ===")
	for i, data := range testData {
		proof, err := tree.GenerateProof(leaves[i])
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s's Proof:\n", data.name)
		fmt.Printf("  Address: %s\n", data.address.Hex())
		fmt.Printf("  Amount:  %s wei\n", data.amount.String())
		fmt.Printf("  Leaf:    %s\n", leaves[i].Hex())
		fmt.Printf("  Proof:   [")
		for j, element := range proof {
			if j > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", element.Hex())
		}
		fmt.Printf("]\n")
		fmt.Println()
	}

	// Contract verification format
	fmt.Println("=== STEP 4: Contract Verification Data ===")
	fmt.Println("Use these values in your TokenClaimer.sol contract:")
	fmt.Printf("Root: %s\n", root.Hex())
	fmt.Println()

	for i, data := range testData {
		proof, _ := tree.GenerateProof(leaves[i])
		fmt.Printf("Test %d (%s):\n", i+1, data.name)
		fmt.Printf("  verifyAddress([")
		for j, element := range proof {
			if j > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", element.Hex())
		}
		fmt.Printf("], %s, %s, %s)\n", root.Hex(), data.address.Hex(), data.amount.String())
		fmt.Println()
	}

	// Verify all proofs
	fmt.Println("=== STEP 5: Verification Test ===")
	allValid := true
	for i, data := range testData {
		proof, _ := tree.GenerateProof(leaves[i])
		isValid := merkle.VerifyProof(proof, root, leaves[i])
		fmt.Printf("%s: %t\n", data.name, isValid)
		if !isValid {
			allValid = false
		}
	}
	fmt.Printf("\nAll proofs valid: %t\n", allValid)

	fmt.Println("\n=== Test completed! ===")
} 