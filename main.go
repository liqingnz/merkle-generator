package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"merkle-generator/merkle"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "merkle-generator",
	Short: "A CLI tool for generating Merkle trees, roots, and proofs",
	Long:  `A CLI tool that generates Merkle tree roots and proofs for given bytes32 leaves.`,
}

var generateRootCmd = &cobra.Command{
	Use:   "root [leaf1] [leaf2] ...",
	Short: "Generate Merkle root from leaves",
	Long:  `Generate Merkle root from a list of hex-encoded bytes32 leaves.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		leaves, err := parseLeaves(args)
		if err != nil {
			fmt.Printf("Error parsing leaves: %v\n", err)
			os.Exit(1)
		}

		tree, err := merkle.NewMerkleTree(leaves)
		if err != nil {
			fmt.Printf("Error creating Merkle tree: %v\n", err)
			os.Exit(1)
		}

		root := tree.GenerateRoot()
		fmt.Printf("Merkle Root: %s\n", root.Hex())
	},
}

var generateProofCmd = &cobra.Command{
	Use:   "proof [target] [leaf1] [leaf2] ...",
	Short: "Generate Merkle proof for a target leaf",
	Long:  `Generate Merkle proof for a target leaf given a list of all leaves.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse target the same way as leaves
		targetLeaves, err := parseLeaves([]string{args[0]})
		if err != nil {
			fmt.Printf("Error parsing target: %v\n", err)
			os.Exit(1)
		}
		target := targetLeaves[0]

		leaves, err := parseLeaves(args[1:])
		if err != nil {
			fmt.Printf("Error parsing leaves: %v\n", err)
			os.Exit(1)
		}

		tree, err := merkle.NewMerkleTree(leaves)
		if err != nil {
			fmt.Printf("Error creating Merkle tree: %v\n", err)
			os.Exit(1)
		}

		proof, err := tree.GenerateProof(target)
		if err != nil {
			fmt.Printf("Error generating proof: %v\n", err)
			os.Exit(1)
		}

		root := tree.GenerateRoot()

		// Output as JSON for easy integration
		result := map[string]interface{}{
			"target": target.Hex(),
			"root":   root.Hex(),
			"proof":  formatProof(proof),
		}

		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
	},
}

var verifyProofCmd = &cobra.Command{
	Use:   "verify [root] [target] [proof1] [proof2] ...",
	Short: "Verify a Merkle proof",
	Long:  `Verify if a target leaf is in the Merkle tree using the provided proof.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse root and target the same way as leaves
		rootLeaves, err := parseLeaves([]string{args[0]})
		if err != nil {
			fmt.Printf("Error parsing root: %v\n", err)
			os.Exit(1)
		}
		root := rootLeaves[0]

		targetLeaves, err := parseLeaves([]string{args[1]})
		if err != nil {
			fmt.Printf("Error parsing target: %v\n", err)
			os.Exit(1)
		}
		target := targetLeaves[0]

		var proof []common.Hash
		if len(args) > 2 {
			proof, err = parseLeaves(args[2:])
			if err != nil {
				fmt.Printf("Error parsing proof: %v\n", err)
				os.Exit(1)
			}
		}

		isValid := merkle.VerifyProof(proof, root, target)
		fmt.Printf("Proof is valid: %t\n", isValid)
	},
}

var hashDataCmd = &cobra.Command{
	Use:   "hash [data]",
	Short: "Hash arbitrary data to bytes32",
	Long:  `Hash arbitrary string data to bytes32 using Keccak256.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash := merkle.HashData([]byte(args[0]))
		fmt.Printf("Hash: %s\n", hash.Hex())
	},
}

var hashAddressAmountCmd = &cobra.Command{
	Use:   "hash-address-amount [address] [amount]",
	Short: "Hash address + amount to bytes32 (like TokenClaimer.sol)",
	Long:  `Hash address + amount to bytes32 using keccak256(abi.encodePacked(address, amount)).`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		address := common.HexToAddress(args[0])
		if address == (common.Address{}) {
			fmt.Printf("Error parsing address: %s\n", args[0])
			os.Exit(1)
		}

		amount, ok := new(big.Int).SetString(args[1], 10)
		if !ok {
			fmt.Printf("Error parsing amount: %s\n", args[1])
			os.Exit(1)
		}

		hash := merkle.HashAddressAmount(address, amount)
		fmt.Printf("Address: %s\n", address.Hex())
		fmt.Printf("Amount: %s\n", amount.String())
		fmt.Printf("Hash: %s\n", hash.Hex())
	},
}

func parseLeaves(args []string) ([]common.Hash, error) {
	leaves := make([]common.Hash, len(args))
	for i, arg := range args {
		// Remove any whitespace
		arg = strings.TrimSpace(arg)

		// If it doesn't start with 0x, assume it's raw data to be hashed
		if !strings.HasPrefix(arg, "0x") {
			leaves[i] = merkle.HashData([]byte(arg))
		} else {
			hash, err := merkle.HexToHash(arg)
			if err != nil {
				return nil, fmt.Errorf("invalid hex at position %d: %v", i, err)
			}
			leaves[i] = hash
		}
	}
	return leaves, nil
}

func formatProof(proof []common.Hash) []string {
	result := make([]string, len(proof))
	for i, hash := range proof {
		result[i] = hash.Hex()
	}
	return result
}

func init() {
	rootCmd.AddCommand(generateRootCmd)
	rootCmd.AddCommand(generateProofCmd)
	rootCmd.AddCommand(verifyProofCmd)
	rootCmd.AddCommand(hashDataCmd)
	rootCmd.AddCommand(hashAddressAmountCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
