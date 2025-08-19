// Package util provides CSV reading utilities for claimer data
package util

import (
	"encoding/csv"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

// ClaimerInfo contains information about a claimer from CSV
type ClaimerInfo struct {
	Name       string
	Address    common.Address
	PrivateKey string
	Amount     *big.Int
}

// ReadClaimersFromCSV reads claimer information from a CSV file
// CSV format: address,private_key,amount
func ReadClaimersFromCSV(filePath string) ([]ClaimerInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	claimers := make([]ClaimerInfo, len(records))
	for i, record := range records {
		if len(record) != 3 {
			return nil, fmt.Errorf("invalid CSV format at line %d: expected 3 columns (address,private_key,amount), got %d", i+1, len(record))
		}

		// Parse address
		address := common.HexToAddress(record[0])

		// Get private key
		privateKey := record[1]

		// Parse amount
		amount, ok := new(big.Int).SetString(record[2], 10)
		if !ok {
			return nil, fmt.Errorf("invalid amount at line %d: %s", i+1, record[2])
		}

		claimers[i] = ClaimerInfo{
			Name:       fmt.Sprintf("Claimer_%d", i+1),
			Address:    address,
			PrivateKey: privateKey,
			Amount:     amount,
		}
	}

	return claimers, nil
}

// ReadCSVTestCases reads test cases from claimer info (for merkle tree generation)
func ReadCSVTestCases(filePath string) ([]TestCase, error) {
	claimers, err := ReadClaimersFromCSV(filePath)
	if err != nil {
		return nil, err
	}

	testCases := make([]TestCase, len(claimers))
	for i, claimer := range claimers {
		testCases[i] = TestCase{
			Name:    claimer.Name,
			Address: claimer.Address,
			Amount:  claimer.Amount,
		}
	}

	return testCases, nil
}

// ReadCSVAddressesAndAmounts reads addresses and amounts from CSV file
// Returns separate slices for addresses and amounts
func ReadCSVAddressesAndAmounts(filePath string) ([]common.Address, []*big.Int, error) {
	testCases, err := ReadCSVTestCases(filePath)
	if err != nil {
		return nil, nil, err
	}

	addresses := make([]common.Address, len(testCases))
	amounts := make([]*big.Int, len(testCases))

	for i, testCase := range testCases {
		addresses[i] = testCase.Address
		amounts[i] = testCase.Amount
	}

	return addresses, amounts, nil
}
