# Util Package

This package provides reusable utilities for Ethereum interaction and TokenClaimer contract operations.

## Modules

### ethclient.go - Ethereum Client Utilities

Provides enhanced Ethereum client functionality with utilities for:

- **EthClient**: Wrapper around ethclient.Client with cached chain ID
- **AccountInfo**: Account management with private key, address, and nonce
- **TransactionParams**: Standardized transaction parameter structure
- **Transaction utilities**: Gas estimation, transaction creation, signing, and sending

**Key Functions:**

- `NewEthClient(rpcURL)` - Create enhanced Ethereum client
- `GetAccountInfo(privateKey)` - Derive account info from private key
- `EstimateAndCreateTransaction()` - Create transactions with gas estimation
- `SignAndSendTransaction()` - Sign and broadcast transactions

### contract.go - TokenClaimer Contract Utilities

Provides high-level interface for TokenClaimer contract interactions:

- **TokenClaimerContract**: Wrapper for contract operations
- **TestCase**: Standardized address/amount data structure
- **MerkleData**: Merkle tree data management
- **Contract functions**: Root generation, proof generation, address verification

**Key Functions:**

- `NewTokenClaimerContract()` - Create contract wrapper
- `GenerateMerkleRoot()` - Call contract's root generation
- `GenerateProof()` - Generate proofs via contract
- `VerifyAddress()` - Verify addresses using contract
- `PrepareClaimTransaction()` - Prepare claim transaction data
- `GenerateLocalMerkleData()` - Generate local merkle trees
- `FindTestCaseIndex()` - Find test case by address

### config.go - Configuration Management

Provides configuration file handling and validation for simplified CSV-based workflow:

- **Config**: Main configuration structure with RPC and CSV settings
- **RPCConfig**: RPC connection settings
- **CSVConfig**: CSV file path configuration

**Key Functions:**

- `LoadConfig(filename)` - Load YAML configuration
- `ValidateConfig()` - Validate configuration completeness

### csv.go - CSV Data Processing

Provides utilities for reading claimer data from CSV files:

- **ClaimerInfo**: Individual claimer information (address, private key, amount)
- **ReadClaimersFromCSV()** - Read complete claimer data from CSV
- **ReadCSVTestCases()** - Convert claimer data to test cases for merkle tree
- **ReadCSVAddressesAndAmounts()** - Read and separate addresses/amounts

**CSV Format:**

```csv
address,private_key,amount
0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6,0x123abc...,1000000000000000000
0x8ba1f109551bD432803012645Hac136c30F7D00E,0x456def...,2000000000000000000
```

## Usage

```go
import "merkle-generator/util"

// Load configuration
config, err := util.LoadConfig("config.yml")

// Connect to Ethereum
client, err := util.NewEthClient(config.RPC.Endpoint)

// Create contract wrapper
contract, err := util.NewTokenClaimerContract(config.RPC.ContractAddress, client)

// Read claimer info from CSV file
claimers, err := util.ReadClaimersFromCSV(config.CSV.FilePath)

// Get account info for first claimer
account, err := client.GetAccountInfo(claimers[0].PrivateKey)

// Read test cases from CSV file (for merkle tree generation)
testCases, err := util.ReadCSVTestCases(config.CSV.FilePath)

// Generate merkle data
merkleData, err := util.GenerateLocalMerkleData(testCases)

// Generate proof for specific claimer
proof, err := merkleData.GenerateLocalProof(0) // Index 0 for first claimer

// Prepare claim transaction for individual claim
data, err := contract.PrepareClaimTransaction(claimers[0].Address, claimers[0].Amount, proof)
```

## Design Principles

1. **Separation of Concerns**: Each module handles a specific aspect of operations
2. **Error Handling**: Comprehensive error wrapping with context
3. **Reusability**: Functions designed for use across multiple tools
4. **Type Safety**: Strong typing with custom structs
5. **Testing**: Designed to be easily testable and mockable
