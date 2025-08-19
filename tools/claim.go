// Package tools provides utilities for interacting with TokenClaimer contracts
//
// This package includes tools for:
// - Generating proofs for token claims
// - Sending claim transactions to contracts
// - Verifying merkle proofs
//
// Usage:
//
//	go run tools/claim.go -config config.yml -address 0x... -amount 1000
package main

import (
	"flag"
	"fmt"
	"log"

	"merkle-generator/merkle"
	"merkle-generator/util"

	"github.com/ethereum/go-ethereum/common"
)

// Command line flags
var (
	configFile = flag.String("config", "examples/config.yml", "Path to configuration file")
	targetAddr = flag.String("address", "", "Target address to claim for (optional - claims for all addresses if not specified)")
	dryRun     = flag.Bool("dry-run", false, "Generate proof and prepare transaction but don't send it")
	help       = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	fmt.Println("=== TokenClaimer Claim Tool ===")

	// Load configuration
	config, err := util.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := config.ValidateConfig(); err != nil {
		fmt.Printf("⚠️  Configuration error: %v\n", err)
		fmt.Println("   Please update config file with your settings:")
		fmt.Println("   - rpc.endpoint: Your Ethereum RPC URL")
		fmt.Println("   - rpc.contract_address: Your TokenClaimer contract address")
		fmt.Println("   - rpc.private_key: Your private key for transactions")
		return
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

	fmt.Printf("Contract address: %s\n\n", contract.Address.Hex())

	// Execute claim process
	err = executeClaims(contract, config)
	if err != nil {
		log.Fatalf("Claim failed: %v", err)
	}
}

func showHelp() {
	fmt.Println("TokenClaimer Claim Tool")
	fmt.Println()
	fmt.Println("This tool helps you claim tokens from a TokenClaimer contract by:")
	fmt.Println("1. Generating the merkle proof for your address")
	fmt.Println("2. Creating and signing the claim transaction")
	fmt.Println("3. Sending the transaction to the network")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run tools/claim.go [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Claim for all addresses in CSV")
	fmt.Println("  go run tools/claim.go -config examples/config.yml")
	fmt.Println()
	fmt.Println("  # Dry run to see what would happen")
	fmt.Println("  go run tools/claim.go -config examples/config.yml -dry-run")
	fmt.Println()
	fmt.Println("  # Claim for specific address")
	fmt.Println("  go run tools/claim.go -config examples/config.yml -address 0x123...")
}

func executeClaims(contract *util.TokenClaimerContract, config *util.Config) error {
	// Read claimer information from CSV file
	fmt.Printf("Reading claimer data from CSV: %s\n", config.CSV.FilePath)
	claimers, err := util.ReadClaimersFromCSV(config.CSV.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read claimers from CSV: %v", err)
	}
	fmt.Printf("Loaded %d claimers from CSV\n", len(claimers))

	// Filter claimers based on target address if specified
	var targetClaimers []util.ClaimerInfo
	if *targetAddr != "" {
		targetAddress := common.HexToAddress(*targetAddr)
		for _, claimer := range claimers {
			if claimer.Address == targetAddress {
				targetClaimers = append(targetClaimers, claimer)
				break
			}
		}
		if len(targetClaimers) == 0 {
			return fmt.Errorf("address %s not found in CSV", targetAddress.Hex())
		}
	} else {
		// Claim for all addresses in CSV
		targetClaimers = claimers
	}

	fmt.Printf("Will process %d claim(s)\n", len(targetClaimers))

	// Generate merkle tree data from all claimers
	fmt.Println("\n=== Generating Merkle Tree ===")
	testCases, err := util.ReadCSVTestCases(config.CSV.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read test cases: %v", err)
	}

	merkleData, err := util.GenerateLocalMerkleData(testCases)
	if err != nil {
		return fmt.Errorf("failed to generate merkle data: %v", err)
	}
	fmt.Printf("Merkle Root: %s\n", merkleData.Root.Hex())

	// Process each claim individually
	for i, claimer := range targetClaimers {
		fmt.Printf("\n=== Processing Claim %d/%d ===\n", i+1, len(targetClaimers))
		fmt.Printf("Claimer: %s (%s)\n", claimer.Name, claimer.Address.Hex())
		fmt.Printf("Amount: %s\n", claimer.Amount.String())

		// Find the index for this claimer in the test cases
		targetIndex := util.FindTestCaseIndex(testCases, claimer.Address)
		if targetIndex == -1 {
			fmt.Printf("⚠️  Skipping %s - not found in merkle tree data\n", claimer.Address.Hex())
			continue
		}

		// Generate proof for this claimer
		proof, err := merkleData.GenerateLocalProof(targetIndex)
		if err != nil {
			fmt.Printf("❌ Failed to generate proof for %s: %v\n", claimer.Address.Hex(), err)
			continue
		}

		// Verify proof locally
		targetLeaf := merkle.HashAddressAmount(claimer.Address, claimer.Amount)
		isValid := merkle.VerifyProof(proof, merkleData.Root, targetLeaf)
		if !isValid {
			fmt.Printf("❌ Invalid proof for %s\n", claimer.Address.Hex())
			continue
		}
		fmt.Printf("✅ Proof verification successful\n")

		if *dryRun {
			fmt.Printf("🏃 Dry run - would claim %s for %s\n", claimer.Amount.String(), claimer.Address.Hex())
			continue
		}

		// Execute the claim
		err = executeSingleClaim(contract, claimer, proof, merkleData.Root)
		if err != nil {
			fmt.Printf("❌ Failed to claim for %s: %v\n", claimer.Address.Hex(), err)
			continue
		}

		fmt.Printf("✅ Successfully claimed for %s\n", claimer.Address.Hex())
	}

	fmt.Println("\n=== Claims Complete ===")
	return nil
}

func executeSingleClaim(contract *util.TokenClaimerContract, claimer util.ClaimerInfo, proof []common.Hash, merkleRoot common.Hash) error {
	// Prepare claim transaction
	data, err := contract.PrepareClaimTransaction(claimer.Address, claimer.Amount, proof)
	if err != nil {
		return fmt.Errorf("failed to prepare claim transaction: %v", err)
	}

	// Get account info for this claimer
	account, err := contract.Client.GetAccountInfo(claimer.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to get account info for claimer: %v", err)
	}

	// Create transaction parameters
	txParams := util.TransactionParams{
		To:   contract.Address,
		Data: data,
	}

	// Create transaction
	tx, err := contract.Client.EstimateAndCreateTransaction(account, txParams)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	fmt.Printf("  Transaction details:\n")
	fmt.Printf("    From: %s\n", account.Address.Hex())
	fmt.Printf("    To (contract): %s\n", contract.Address.Hex())
	fmt.Printf("    Gas Limit: %d\n", tx.Gas())
	fmt.Printf("    Gas Price: %s wei\n", tx.GasPrice().String())
	fmt.Printf("    Nonce: %d\n", tx.Nonce())

	// Sign and send transaction
	signedTx, err := contract.Client.SignAndSendTransaction(tx, account.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign and send transaction: %v", err)
	}

	fmt.Printf("  Transaction sent: %s\n", signedTx.Hash().Hex())

	return nil
}
