// Package main provides an example of testing TokenClaimer contract read functions via RPC
//
// This example demonstrates:
// 1. Connecting to an Ethereum node via RPC
// 2. Testing contract's generateMerkleRoot function
// 3. Testing contract's generateProof function
// 4. Testing contract's verifyAddress function
//
// For claim functionality (sending transactions), use the separate tools/claim.go
//
// To run this example:
// 1. Copy examples/config.yml.example to examples/config.yml
// 2. Update config.yml with your RPC endpoint and contract address
// 3. Uncomment the main function at the bottom of this file
// 4. Run: go run examples/rpc_contract_tester.go
package main

import (
	"fmt"
	"log"

	"merkle-generator/merkle"
	"merkle-generator/util"

	"github.com/ethereum/go-ethereum/common"
)

func runContractTester() {
	fmt.Println("=== TokenClaimer Contract RPC Test (with config) ===")

	// Load configuration
	config, err := util.LoadConfig("examples/config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := config.ValidateConfig(); err != nil {
		fmt.Printf("⚠️  Configuration error: %v\n", err)
		fmt.Println("   Please update config.yml with your settings:")
		fmt.Println("   - rpc.endpoint: Your Ethereum RPC URL")
		fmt.Println("   - rpc.contract_address: Your TokenClaimer contract address")
		return
	}

	// Convert config to test cases
	testCases, err := config.ToTestCases()
	if err != nil {
		log.Fatalf("Failed to convert config to test cases: %v", err)
	}

	// Connect to Ethereum node
	client, err := util.NewEthClient(config.RPC.Endpoint)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	fmt.Printf("Connected to chain ID: %s\n", client.ChainID.String())

	// Create contract wrapper
	contract, err := util.NewTokenClaimerContract(config.RPC.ContractAddress, client)
	if err != nil {
		log.Fatalf("Failed to create contract wrapper: %v", err)
	}

	fmt.Printf("Testing contract at: %s\n\n", contract.Address.Hex())

	// Test contract's generation functions and get calculated data
	contractRoot, localRoot, contractProofs, localProofs, leaves := runContractGenerationTest(contract, testCases)

	// Use the calculated data for verification testing
	runContractTest(contract, testCases, contractRoot, localRoot, contractProofs, localProofs, leaves)
}

func runContractGenerationTest(contract *util.TokenClaimerContract, testCases []util.TestCase) (common.Hash, common.Hash, [][]common.Hash, [][]common.Hash, []common.Hash) {
	fmt.Println("\n=== STEP 3: Testing Contract Generation Functions ===")

	// Test generateMerkleRoot
	fmt.Println("Testing generateMerkleRoot...")
	contractRoot, err := contract.GenerateMerkleRoot(testCases)
	if err != nil {
		log.Printf("Failed to call generateMerkleRoot: %v", err)
		return common.Hash{}, common.Hash{}, nil, nil, nil
	}

	// Generate our local merkle data for comparison
	merkleData, err := util.GenerateLocalMerkleData(testCases)
	if err != nil {
		log.Printf("Failed to generate local merkle data: %v", err)
		return common.Hash{}, common.Hash{}, nil, nil, nil
	}

	// Initialize proof arrays
	contractProofs := make([][]common.Hash, len(testCases))
	localProofs := make([][]common.Hash, len(testCases))

	fmt.Printf("Contract Root: %s\n", contractRoot.Hex())
	fmt.Printf("Local Root:    %s\n", merkleData.Root.Hex())

	if contractRoot == merkleData.Root {
		fmt.Printf("✅ Contract and local root generation match!\n")
	} else {
		fmt.Printf("❌ Contract and local root generation differ!\n")
	}

	// Test generateProof for each address
	fmt.Println("\nTesting generateProof for each address...")
	for i, testCase := range testCases {
		fmt.Printf("Testing generateProof for %s...\n", testCase.Name)

		// Get contract proof
		contractProof, err := contract.GenerateProof(testCases, testCase.Address)
		if err != nil {
			log.Printf("Failed to call generateProof for %s: %v", testCase.Name, err)
			continue
		}

		// Generate our local proof for comparison
		localProof, err := merkleData.GenerateLocalProof(i)
		if err != nil {
			log.Printf("Failed to generate local proof for %s: %v", testCase.Name, err)
			continue
		}

		// Store proofs for reuse
		contractProofs[i] = contractProof
		localProofs[i] = localProof

		fmt.Printf("  Contract Proof: [")
		for j, p := range contractProof {
			if j > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", p.Hex())
		}
		fmt.Printf("]\n")

		fmt.Printf("  Local Proof:    [")
		for j, p := range localProof {
			if j > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", p.Hex())
		}
		fmt.Printf("]\n")

		// Compare proofs
		if len(contractProof) == len(localProof) {
			match := true
			for j := range contractProof {
				if contractProof[j] != localProof[j] {
					match = false
					break
				}
			}
			if match {
				fmt.Printf("  ✅ Contract and local proof generation match!\n")
			} else {
				fmt.Printf("  ❌ Contract and local proof generation differ!\n")
			}
		} else {
			fmt.Printf("  ❌ Proof lengths differ: contract=%d, local=%d\n", len(contractProof), len(localProof))
		}
		fmt.Println()
	}

	// Return all calculated data for reuse
	return contractRoot, merkleData.Root, contractProofs, localProofs, merkleData.Leaves
}

func runContractTest(contract *util.TokenClaimerContract, testCases []util.TestCase, contractRoot common.Hash, localRoot common.Hash, contractProofs [][]common.Hash, localProofs [][]common.Hash, leaves []common.Hash) {
	// Display the pre-calculated data
	fmt.Println("=== STEP 1: Using Pre-calculated Data ===")
	for i, testCase := range testCases {
		fmt.Printf("%s: %s (addr: %s, amount: %s)\n",
			testCase.Name, leaves[i].Hex(), testCase.Address.Hex(), testCase.Amount.String())
	}

	fmt.Printf("\nMerkle Root: %s\n\n", localRoot.Hex())

	// Test each case with the contract using pre-calculated proofs
	fmt.Println("=== STEP 2: Testing Contract Calls (using pre-calculated proofs) ===")
	for i, testCase := range testCases {
		fmt.Printf("Testing %s...\n", testCase.Name)

		// Use pre-calculated proof
		proof := localProofs[i]

		// Call contract verification
		isValid, err := contract.VerifyAddress(proof, localRoot, testCase.Address, testCase.Amount)
		if err != nil {
			log.Printf("Failed to call contract verifyAddress for %s: %v", testCase.Name, err)
			continue
		}

		// Display result
		status := "❌ INVALID"
		if isValid {
			status = "✅ VALID"
		}
		fmt.Printf("  %s: %s\n", testCase.Name, status)

		// Also verify locally for comparison
		localValid := merkle.VerifyProof(proof, localRoot, leaves[i])
		localStatus := "❌ INVALID"
		if localValid {
			localStatus = "✅ VALID"
		}
		fmt.Printf("  Local verification: %s\n", localStatus)

		if isValid == localValid {
			fmt.Printf("  ✅ Contract and local verification match!\n")
		} else {
			fmt.Printf("  ❌ Mismatch between contract and local verification!\n")
		}
		fmt.Println()
	}
}

// Uncomment the main function below and run this file directly to test:
// go run examples/rpc_contract_tester.go
/*
func main() {
	runContractTester()
}
*/
