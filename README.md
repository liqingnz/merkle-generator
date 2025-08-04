# Merkle Tree Generator

A Go CLI tool for generating Merkle trees, roots, and proofs for bytes32 leaves. This implementation follows the same logic as the Solidity MerkleTree library for consistency.

## Features

- Generate Merkle root from a list of leaves
- Generate Merkle proof for a specific leaf
- Verify Merkle proofs
- Hash arbitrary data to bytes32
- CLI interface for easy integration
- Compatible with Solidity MerkleTree implementation

## Installation

```bash
cd merkleTree/merkle-generator
go mod tidy
go build -o merkle-generator .
```

## Usage

### Generate Merkle Root

Generate a Merkle root from hex-encoded bytes32 leaves:

```bash
./merkle-generator root 0x1234... 0x5678... 0x9abc...
```

Or from raw string data (will be hashed automatically):

```bash
./merkle-generator root alice bob charlie
```

### Generate Merkle Proof

Generate a proof for a specific target leaf:

```bash
./merkle-generator proof alice alice bob charlie
```

Output will be JSON format:

```json
{
  "target": "0x...",
  "root": "0x...",
  "proof": ["0x...", "0x..."]
}
```

### Verify Merkle Proof

Verify if a proof is valid:

```bash
./merkle-generator verify <root> <target> <proof1> <proof2> ...
```

### Hash Data

Hash arbitrary string data to bytes32:

```bash
./merkle-generator hash "alice"
```

### Hash Address + Amount

Hash address + amount to bytes32 (compatible with TokenClaimer.sol):

```bash
./merkle-generator hash-address-amount 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 1000000000000000000
# Output: 0x862d9f69cd1642f07c56ec6b92856ce141af9dfe404d2ab0c4685a334945ffe6
```

**Note**: This matches Solidity's `keccak256(abi.encodePacked(address, uint256))` exactly.

## Examples

### Basic Example

```bash
# Generate root for three users
./merkle-generator root alice bob charlie

# Generate proof for alice
./merkle-generator proof alice alice bob charlie

# Verify the proof (using the root and proof from above commands)
./merkle-generator verify <root> <alice_hash> <proof_elements>
```

### Working with Hex Values

```bash
# Using hex-encoded values
./merkle-generator root 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef 0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321

# Generate proof for the first leaf
./merkle-generator proof 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef 0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321
```

### Working with Address + Amount (TokenClaimer.sol compatible)

```bash
# Generate leaves from address + amount combinations
./merkle-generator hash-address-amount 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 1000000000000000000

# Create Merkle tree from address+amount leaves
./merkle-generator root 0xf73127d6a4e0e3f816c74879ecafe4cff07a19760dc043d11fad3ac9de93a96e 0xd81cec089f0d58b327c1c3bb5280428f6b81e3b908325dbe2cadc941751d2581

# Generate proof for specific address+amount
./merkle-generator proof 0xf73127d6a4e0e3f816c74879ecafe4cff07a19760dc043d11fad3ac9de93a96e 0xf73127d6a4e0e3f816c74879ecafe4cff07a19760dc043d11fad3ac9de93a96e 0xd81cec089f0d58b327c1c3bb5280428f6b81e3b908325dbe2cadc941751d2581
```

### RPC Contract Testing

Test your deployed TokenClaimer contract directly:

#### Option 1: With Configuration File (Recommended)

```bash
# Edit examples/config.yml with your settings
go run examples/rpc_contract_tester.go
```

#### Option 2: Hardcoded Configuration

```bash
# Edit the file directly
go run examples/rpc_contract_tester.go
```

**Configuration**:

- **examples/config.yml**: Edit this file for easy configuration
- **Hardcoded**: Edit the Go file directly

**Settings**:

- `rpc.endpoint`: Your Ethereum RPC URL (e.g., Infura, Alchemy)
- `rpc.contract_address`: Your deployed TokenClaimer contract address
- `rpc.private_key`: Optional, for transactions
- `test_data.addresses/amounts`: Customize test data

**Setup**:

1. Copy `examples/config.yml.example` to `examples/config.yml`
2. Update `examples/config.yml` with your values
3. Note: `config.yml` is in `.gitignore` to protect sensitive data

**Features**:

1. ✅ **verifyAddress testing**: Tests your contract's verification function
2. ✅ **generateMerkleRoot testing**: Compares contract vs local root generation
3. ✅ **generateProof testing**: Compares contract vs local proof generation
4. ✅ **Optimized workflow**: Reuses calculated root/proofs to avoid duplication
5. ✅ **Full compatibility verification**: Ensures Go and Solidity match exactly
6. ✅ **Detailed output**: Shows all hashes, proofs, and comparison results

## Integration

### As a Library

```go
package main

import (
    "fmt"
    "merkle-generator/merkle"
    "github.com/ethereum/go-ethereum/common"
)

func main() {
    // Create leaves
    leaves := []common.Hash{
        merkle.HashData([]byte("alice")),
        merkle.HashData([]byte("bob")),
        merkle.HashData([]byte("charlie")),
    }

    // Create tree
    tree, err := merkle.NewMerkleTree(leaves)
    if err != nil {
        panic(err)
    }

    // Generate root
    root := tree.GenerateRoot()
    fmt.Printf("Root: %s\n", root.Hex())

    	// Generate proof for alice
	proof, err := tree.GenerateProof(leaves[0])
	if err != nil {
		panic(err)
	}

	// Verify proof
	isValid := merkle.VerifyProof(proof, root, leaves[0])
	fmt.Printf("Proof valid: %t\n", isValid)
}

// Example with address + amount (TokenClaimer.sol compatible)
func exampleWithAddressAmount() {
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")
	amount := big.NewInt(1000000000000000000) // 1 ETH in wei

	// Create leaf hash like Solidity: keccak256(abi.encodePacked(address, amount))
	leaf := merkle.HashAddressAmount(address, amount)

	leaves := []common.Hash{leaf}
	tree, _ := merkle.NewMerkleTree(leaves)
	root := tree.GenerateRoot()

	fmt.Printf("Address: %s\n", address.Hex())
	fmt.Printf("Amount: %s\n", amount.String())
	fmt.Printf("Leaf: %s\n", leaf.Hex())
	fmt.Printf("Root: %s\n", root.Hex())
}
```

## Testing

Run the test suite:

```bash
go test ./merkle -v
```

## Compatibility

This implementation uses the same hashing and tree construction logic as the Solidity MerkleTree library:

- Uses Keccak256 for hashing
- Maintains consistent ordering in hash pairs (a < b)
- Handles odd numbers of leaves by promoting the last leaf
- Compatible proof format with OpenZeppelin's MerkleProof library
- **Address+Amount hashing**: Exactly matches Solidity's `keccak256(abi.encodePacked(address, uint256))`
  - Address: 20 bytes
  - Amount: 32 bytes (uint256, big-endian)
  - Total: 52 bytes concatenated before hashing

### Verified Compatibility

The `HashAddressAmount` function has been tested to produce identical results to Solidity:

```go
// Go
address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")
amount := big.NewInt(1000000000000000000) // 1 ETH
hash := merkle.HashAddressAmount(address, amount)
// Result: 0x862d9f69cd1642f07c56ec6b92856ce141af9dfe404d2ab0c4685a334945ffe6
```

```solidity
// Solidity
function leaf(address _addr, uint256 _amount) public pure returns (bytes32) {
    return keccak256(abi.encodePacked(_addr, _amount));
}
// leaf(0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6, 1000000000000000000)
// Result: 0x862d9f69cd1642f07c56ec6b92856ce141af9dfe404d2ab0c4685a334945ffe6
```

## Dependencies

- `github.com/ethereum/go-ethereum` - For common.Hash types and Keccak256 hashing
- `github.com/spf13/cobra` - For CLI interface
