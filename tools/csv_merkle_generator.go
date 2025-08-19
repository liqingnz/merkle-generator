package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"os"

	"merkle-generator/merkle"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: go run tools/csv_merkle_generator.go <csv_file> [verbose]")
		fmt.Println("Example: go run tools/csv_merkle_generator.go data/airdrop.csv")
		fmt.Println("Example: go run tools/csv_merkle_generator.go data/airdrop.csv true")
		os.Exit(1)
	}

	csvFile := os.Args[1]
	verbose := false
	if len(os.Args) == 3 && os.Args[2] == "true" {
		verbose = true
	}

	// Read CSV file
	addresses, amounts, err := readCSV(csvFile)
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	fmt.Printf("=== Processing CSV: %s ===\n", csvFile)
	fmt.Printf("Total entries: %d\n\n", len(addresses))

	// Generate leaves from addresses and amounts
	leaves := make([]common.Hash, len(addresses))
	for i := range addresses {
		leaves[i] = merkle.HashAddressAmount(addresses[i], amounts[i])
		if verbose || i < 5 || i == len(addresses)-1 {
			fmt.Printf("Entry %d: %s (amount: %s) -> Leaf: %s\n",
				i+1, addresses[i].Hex(), amounts[i].String(), leaves[i].Hex())
		} else if i == 5 {
			fmt.Printf("... (showing first 5 and last entry, use 'true' as second argument for verbose output)\n")
		}
	}

	// Create Merkle tree
	tree, err := merkle.NewMerkleTree(leaves)
	if err != nil {
		log.Fatalf("Error creating Merkle tree: %v", err)
	}

	// Generate root
	root := tree.GenerateRoot()
	fmt.Printf("\n=== Merkle Root ===\n")
	fmt.Printf("Root: %s\n\n", root.Hex())

	// Generate proof for the first entry (as requested)
	if len(leaves) > 0 {
		fmt.Printf("=== Proof for First Entry ===\n")
		firstAddress := addresses[0]
		firstAmount := amounts[0]
		firstLeaf := leaves[0]

		proof, err := tree.GenerateProof(firstLeaf)
		if err != nil {
			log.Fatalf("Error generating proof: %v", err)
		}

		fmt.Printf("Address: %s\n", firstAddress.Hex())
		fmt.Printf("Amount: %s\n", firstAmount.String())
		fmt.Printf("Leaf: %s\n", firstLeaf.Hex())
		fmt.Printf("Proof: [")
		for i, p := range proof {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", p.Hex())
		}
		fmt.Printf("]\n")

		// Verify the proof
		isValid := merkle.VerifyProof(proof, root, firstLeaf)
		fmt.Printf("Proof verification: %t\n\n", isValid)

		// Solidity function call format
		fmt.Printf("=== Solidity Contract Call ===\n")
		fmt.Printf("verifyAddress([")
		for i, p := range proof {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", p.Hex())
		}
		fmt.Printf("], %s, %s, %s)\n", root.Hex(), firstAddress.Hex(), firstAmount.String())
	}

	// Save results to files
	saveResults(addresses, amounts, leaves, root, csvFile)

	// Verify a random entry at the end
	if len(leaves) > 100 {
		verifyIndex := 100 // Verify entry 100
		fmt.Printf("\n=== Verification Test ===\n")
		fmt.Printf("Verifying entry %d: %s (amount: %s)\n",
			verifyIndex+1, addresses[verifyIndex].Hex(), amounts[verifyIndex].String())

		verifyProof, err := tree.GenerateProof(leaves[verifyIndex])
		if err != nil {
			log.Printf("Error generating verification proof: %v", err)
		} else {
			isValid := merkle.VerifyProof(verifyProof, root, leaves[verifyIndex])
			fmt.Printf("Verification result: %t\n", isValid)
		}
	}
}

func readCSV(filename string) ([]common.Address, []*big.Int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(records) < 2 {
		return nil, nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// Skip header row
	records = records[1:]

	var addresses []common.Address
	var amounts []*big.Int

	for i, record := range records {
		if len(record) < 2 {
			fmt.Printf("Warning: Skipping row %d (insufficient columns)\n", i+2)
			continue
		}

		// Parse address
		address := common.HexToAddress(record[0])
		if address == (common.Address{}) {
			fmt.Printf("Warning: Skipping row %d (invalid address: %s)\n", i+2, record[0])
			continue
		}

		// Parse amount
		amount, ok := new(big.Int).SetString(record[1], 10)
		if !ok {
			fmt.Printf("Warning: Skipping row %d (invalid amount: %s)\n", i+2, record[1])
			continue
		}

		addresses = append(addresses, address)
		amounts = append(amounts, amount)
	}

	if len(addresses) == 0 {
		return nil, nil, fmt.Errorf("no valid entries found in CSV file")
	}

	fmt.Printf("Processed %d valid entries out of %d total rows\n", len(addresses), len(records))
	return addresses, amounts, nil
}

func saveResults(addresses []common.Address, amounts []*big.Int, leaves []common.Hash, root common.Hash, csvFile string) {
	// Save Merkle root
	rootFile := csvFile + ".root"
	err := os.WriteFile(rootFile, []byte(root.Hex()), 0644)
	if err != nil {
		log.Printf("Warning: Could not save root file: %v", err)
	} else {
		fmt.Printf("Merkle root saved to: %s\n", rootFile)
	}

	// For large datasets, only save a sample of leaves to avoid memory issues
	if len(leaves) > 1000 {
		fmt.Printf("Skipping leaves file generation for large dataset (%d entries)\n", len(leaves))
		fmt.Printf("Use the CLI tool to generate specific proofs: ./merkle-generator proof <leaf> <all_leaves...>\n")
	} else {
		// Save all leaves for smaller datasets
		leavesFile := csvFile + ".leaves"
		leavesContent := "address,amount,leaf\n"
		for i := range addresses {
			leavesContent += fmt.Sprintf("%s,%s,%s\n",
				addresses[i].Hex(), amounts[i].String(), leaves[i].Hex())
		}

		err = os.WriteFile(leavesFile, []byte(leavesContent), 0644)
		if err != nil {
			log.Printf("Warning: Could not save leaves file: %v", err)
		} else {
			fmt.Printf("All leaves saved to: %s\n", leavesFile)
		}
	}

	// Save proof for first entry
	if len(leaves) > 0 {
		tree, _ := merkle.NewMerkleTree(leaves)
		proof, _ := tree.GenerateProof(leaves[0])
		proofFile := csvFile + ".proof"
		proofContent := fmt.Sprintf("address,amount,leaf,proof\n")
		proofContent += fmt.Sprintf("%s,%s,%s,\"[",
			addresses[0].Hex(), amounts[0].String(), leaves[0].Hex())

		for i, p := range proof {
			if i > 0 {
				proofContent += ","
			}
			proofContent += p.Hex()
		}
		proofContent += "]\"\n"

		err = os.WriteFile(proofFile, []byte(proofContent), 0644)
		if err != nil {
			log.Printf("Warning: Could not save proof file: %v", err)
		} else {
			fmt.Printf("First entry proof saved to: %s\n", proofFile)
		}
	}
}
