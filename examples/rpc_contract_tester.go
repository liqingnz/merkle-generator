package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"merkle-generator/merkle"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v3"
)

// TokenClaimer ABI - includes all functions we need
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
	}
]`

// Configuration structs
type Config struct {
	RPC      RPCConfig      `yaml:"rpc"`
	TestData TestDataConfig `yaml:"test_data"`
}

type RPCConfig struct {
	Endpoint        string `yaml:"endpoint"`
	ContractAddress string `yaml:"contract_address"`
	PrivateKey      string `yaml:"private_key"`
}

type TestDataConfig struct {
	Addresses []string `yaml:"addresses"`
	Amounts   []string `yaml:"amounts"`
}

// Test data structure
type TestCase struct {
	Name    string
	Address common.Address
	Amount  *big.Int
}

func main() {
	fmt.Println("=== TokenClaimer Contract RPC Test (with config) ===")

	// Load configuration
	config, err := loadConfig("examples/config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if config.RPC.Endpoint == "" || config.RPC.ContractAddress == "" {
		fmt.Println("⚠️  Please update config.yml with your settings:")
		fmt.Println("   - rpc.endpoint: Your Ethereum RPC URL")
		fmt.Println("   - rpc.contract_address: Your TokenClaimer contract address")
		fmt.Println("   - Optionally set rpc.private_key for transactions")
		return
	}

	// Convert config to test cases
	testCases, err := configToTestCases(config)
	if err != nil {
		log.Fatalf("Failed to convert config to test cases: %v", err)
	}

	// Connect to Ethereum node
	client, err := ethclient.Dial(config.RPC.Endpoint)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	// Test the connection
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}
	fmt.Printf("Connected to chain ID: %s\n", chainID.String())

	// Parse contract ABI
	contractABI, err := abi.JSON(strings.NewReader(TokenClaimerABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	// Contract address
	contractAddr := common.HexToAddress(config.RPC.ContractAddress)
	fmt.Printf("Testing contract at: %s\n\n", contractAddr.Hex())

	// Test contract's generation functions and get calculated data
	contractRoot, localRoot, contractProofs, localProofs, leaves := runContractGenerationTest(client, contractABI, contractAddr, testCases)

	// Use the calculated data for verification testing
	runContractTest(client, contractABI, contractAddr, testCases, contractRoot, localRoot, contractProofs, localProofs, leaves)
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

func configToTestCases(config *Config) ([]TestCase, error) {
	if len(config.TestData.Addresses) != len(config.TestData.Amounts) {
		return nil, fmt.Errorf("addresses and amounts arrays must have the same length")
	}

	testCases := make([]TestCase, len(config.TestData.Addresses))
	for i := range config.TestData.Addresses {
		address := common.HexToAddress(config.TestData.Addresses[i])
		amount, ok := new(big.Int).SetString(config.TestData.Amounts[i], 10)
		if !ok {
			return nil, fmt.Errorf("invalid amount at index %d: %s", i, config.TestData.Amounts[i])
		}

		testCases[i] = TestCase{
			Name:    fmt.Sprintf("User_%d", i+1),
			Address: address,
			Amount:  amount,
		}
	}

	return testCases, nil
}

func runContractGenerationTest(client *ethclient.Client, contractABI abi.ABI, contractAddr common.Address, testCases []TestCase) (common.Hash, common.Hash, [][]common.Hash, [][]common.Hash, []common.Hash) {
	fmt.Println("\n=== STEP 3: Testing Contract Generation Functions ===")

	// Prepare arrays for contract calls
	addresses := make([]common.Address, len(testCases))
	amounts := make([]*big.Int, len(testCases))
	for i, testCase := range testCases {
		addresses[i] = testCase.Address
		amounts[i] = testCase.Amount
	}

	// Test generateMerkleRoot
	fmt.Println("Testing generateMerkleRoot...")

	// Pack the generateMerkleRoot call
	rootData, err := contractABI.Pack("generateMerkleRoot", addresses, amounts)
	if err != nil {
		log.Printf("Failed to pack generateMerkleRoot call: %v", err)
		return common.Hash{}, common.Hash{}, nil, nil, nil
	}

	// Call the contract
	rootCallMsg := ethereum.CallMsg{
		To:   &contractAddr,
		Data: rootData,
	}

	rootResult, err := client.CallContract(context.Background(), rootCallMsg, nil)
	if err != nil {
		log.Printf("Failed to call generateMerkleRoot: %v", err)
		return common.Hash{}, common.Hash{}, nil, nil, nil
	}

	// Unpack the result
	var contractRoot common.Hash
	err = contractABI.UnpackIntoInterface(&contractRoot, "generateMerkleRoot", rootResult)
	if err != nil {
		log.Printf("Failed to unpack generateMerkleRoot result: %v", err)
		return common.Hash{}, common.Hash{}, nil, nil, nil
	}

	// Generate our local root for comparison
	leaves := make([]common.Hash, len(testCases))
	for i, testCase := range testCases {
		leaves[i] = merkle.HashAddressAmount(testCase.Address, testCase.Amount)
	}
	tree, _ := merkle.NewMerkleTree(leaves)
	localRoot := tree.GenerateRoot()

	// Initialize proof arrays
	contractProofs := make([][]common.Hash, len(testCases))
	localProofs := make([][]common.Hash, len(testCases))

	fmt.Printf("Contract Root: %s\n", contractRoot.Hex())
	fmt.Printf("Local Root:    %s\n", localRoot.Hex())

	if contractRoot == localRoot {
		fmt.Printf("✅ Contract and local root generation match!\n")
	} else {
		fmt.Printf("❌ Contract and local root generation differ!\n")
	}

	// Test generateProof for each address
	fmt.Println("\nTesting generateProof for each address...")
	for i, testCase := range testCases {
		fmt.Printf("Testing generateProof for %s...\n", testCase.Name)

		// Pack the generateProof call
		proofData, err := contractABI.Pack("generateProof", addresses, amounts, testCase.Address)
		if err != nil {
			log.Printf("Failed to pack generateProof call for %s: %v", testCase.Name, err)
			continue
		}

		// Call the contract
		proofCallMsg := ethereum.CallMsg{
			To:   &contractAddr,
			Data: proofData,
		}

		proofResult, err := client.CallContract(context.Background(), proofCallMsg, nil)
		if err != nil {
			log.Printf("Failed to call generateProof for %s: %v", testCase.Name, err)
			continue
		}

		// Unpack the result
		var contractProof []common.Hash
		err = contractABI.UnpackIntoInterface(&contractProof, "generateProof", proofResult)
		if err != nil {
			log.Printf("Failed to unpack generateProof result for %s: %v", testCase.Name, err)
			continue
		}

		// Generate our local proof for comparison
		localProof, err := tree.GenerateProof(leaves[i])
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
	return contractRoot, localRoot, contractProofs, localProofs, leaves
}

func runContractTest(client *ethclient.Client, contractABI abi.ABI, contractAddr common.Address, testCases []TestCase, contractRoot common.Hash, localRoot common.Hash, contractProofs [][]common.Hash, localProofs [][]common.Hash, leaves []common.Hash) {
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

		// Pack the function call
		data, err := contractABI.Pack("verifyAddress", proof, localRoot, testCase.Address, testCase.Amount)
		if err != nil {
			log.Printf("Failed to pack function call for %s: %v", testCase.Name, err)
			continue
		}

		// Call the contract
		callMsg := ethereum.CallMsg{
			To:   &contractAddr,
			Data: data,
		}

		result, err := client.CallContract(context.Background(), callMsg, nil)
		if err != nil {
			log.Printf("Failed to call contract for %s: %v", testCase.Name, err)
			continue
		}

		// Unpack the result
		var isValid bool
		err = contractABI.UnpackIntoInterface(&isValid, "verifyAddress", result)
		if err != nil {
			log.Printf("Failed to unpack result for %s: %v", testCase.Name, err)
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
