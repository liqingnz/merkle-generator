# Tools Directory

This directory contains command-line tools for interacting with TokenClaimer contracts and managing merkle proofs.

## Available Tools

### claim.go - Token Claiming Tool

A command-line tool for claiming tokens from TokenClaimer contracts with support for both individual and batch claiming.

**Features:**

- Automatically generates merkle proofs for your address
- Creates and signs claim transactions
- **Batch claiming** using Multicall3 for gas efficiency
- Supports dry-run mode for testing
- Flexible configuration options
- Local proof verification before sending

**Usage:**

**Individual Claims:**

```bash
# Basic claim using config file
go run tools/claim.go -config examples/config.yml

# Dry run to see what would happen without sending transaction
go run tools/claim.go -config examples/config.yml -dry-run

# Claim for specific address and amount
go run tools/claim.go -config examples/config.yml -address 0x123... -amount 1000
```

**Additional Examples:**

```bash
# Show help
go run tools/claim.go -help
```

**Individual Claiming Benefits:**

- **Simple and Reliable**: One transaction per claim, easy to track
- **Independent Processing**: Each claim succeeds or fails independently
- **Private Key Security**: Each claim uses its own private key as `msg.sender`
- **Clear Error Handling**: Failed claims don't affect successful ones

**Configuration:**
The tool uses a CSV-based configuration. Create a config file and CSV data file:

**Config File (config.yml):**

```yaml
rpc:
  endpoint: "https://your-rpc-endpoint.com"
  contract_address: "0x..."

csv:
  file_path: "data/claimers.csv" # Path to CSV with claimer data
```

**CSV File Format:**
The CSV file contains claimer information (address, private key, amount):

```csv
0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6,0x123abc...,1000000000000000000
0x8ba1f109551bD432803012645Hac136c30F7D00E,0x456def...,2000000000000000000
0x9Ac64Cc6634C0532925a3b8D4C9db96C4b4d8b6,0x789ghi...,500000000000000000
```

**How It Works:**

1. **Single CSV**: Contains all claimer data (address, private key, amount)
2. **Individual Claims**: Each address claims using its own private key as `msg.sender`
3. **Merkle Tree**: Generated from all CSV entries for proof generation
4. **Sequential Processing**: Claims are sent one by one, not batched

**Security Notes:**

- Never commit private keys to version control
- Use environment variables or secure key management for production
- Always test with small amounts first
- Use dry-run mode to verify transaction details before sending

### csv_merkle_generator.go - CSV Merkle Tree Generator

A tool for generating merkle trees from CSV files containing address and amount data.

**Features:**

- Reads CSV files with address/amount pairs
- Generates merkle tree and root
- Creates individual proofs for each entry
- Outputs results to files with `.root` and `.proof` extensions
- Supports verbose output for debugging

**Usage:**

```bash
# Generate merkle tree from CSV file
go run tools/csv_merkle_generator.go data/airdrop.csv

# Verbose output showing all entries
go run tools/csv_merkle_generator.go data/airdrop.csv true
```

**CSV Format:**
The CSV file should contain two columns:

- First column: Ethereum addresses (with or without 0x prefix)
- Second column: Token amounts (as integers)

Example CSV:

```
0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6,1000000000000000000
0x8ba1f109551bD432803012645Hac136c30F7D00E,2000000000000000000
```

**Output Files:**

- `filename.csv.root` - Contains the merkle root
- `filename.csv.proof` - Contains JSON with all proofs and verification data

## Future Tools

This directory is designed to be extensible. Future tools might include:

- **verify.go** - Standalone proof verification tool
- **generate.go** - Merkle tree and proof generation tool
- **monitor.go** - Contract event monitoring tool
- **batch.go** - Batch claim processing tool

## Development

When adding new tools:

1. Use the same configuration format for consistency
2. Include comprehensive error handling
3. Add dry-run capabilities for destructive operations
4. Follow the existing code structure and patterns
5. Add appropriate documentation and examples
