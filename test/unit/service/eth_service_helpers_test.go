package service_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/service"
	"github.com/LimeChain/Hederium/test/unit/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setupTest(t *testing.T) (*gomock.Controller, *mocks.MockMirrorClient, *zap.Logger) {
	ctrl := gomock.NewController(t)
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	return ctrl, mockClient, logger
}

func TestGetFeeWeibars_Success(t *testing.T) {
	ctrl, mockClient, logger := setupTest(t)
	defer ctrl.Finish()

	expectedGasTinybars := int64(100000)
	mockClient.EXPECT().
		GetNetworkFees().
		Return(expectedGasTinybars, nil)

	s := service.NewEthService(
		nil,
		mockClient,
		logger,
		nil,
		defaultChainId,
	)

	result, errMap := service.GetFeeWeibars(s, "", "") // Should be handled better!
	assert.Nil(t, errMap)

	// Expected weibars = tinybars * 10^8
	expectedWeibars := big.NewInt(expectedGasTinybars)
	expectedWeibars = expectedWeibars.Mul(expectedWeibars, big.NewInt(100000000))
	assert.Equal(t, expectedWeibars, result)
}

func TestGetFeeWeibars_Error(t *testing.T) {
	ctrl, mockClient, logger := setupTest(t)
	defer ctrl.Finish()

	mockClient.EXPECT().
		GetNetworkFees().
		Return(int64(0), fmt.Errorf("network error"))

	s := service.NewEthService(
		nil,
		mockClient,
		logger,
		nil,
		defaultChainId,
	)

	result, errMap := service.GetFeeWeibars(s, "", "") // Should be handled better
	assert.Nil(t, result)
	assert.NotNil(t, errMap)
	assert.Equal(t, -32000, errMap["code"])
	assert.Equal(t, "Failed to fetch gas price", errMap["message"])
}

func TestProcessBlock_Success(t *testing.T) {
	ctrl, mockClient, logger := setupTest(t)
	defer ctrl.Finish()

	block := &domain.BlockResponse{
		Number:       123,
		Hash:         "0x123abc",
		PreviousHash: "0x456def",
		GasUsed:      1000,
		Size:         2000,
		LogsBloom:    "0x0",
		Timestamp: domain.Timestamp{
			From: "1640995200",
		},
	}

	contractResults := []domain.ContractResults{
		{
			Hash:   "0xtx1",
			Result: "SUCCESS",
		},
		{
			Hash:   "0xtx2",
			Result: "WRONG_NONCE", // Should be filtered out
		},
		{
			Hash:   "0xtx3",
			Result: "SUCCESS",
		},
	}

	mockClient.EXPECT().
		GetContractResults(block.Timestamp).
		Return(contractResults)

	s := service.NewEthService(
		nil,
		mockClient,
		logger,
		nil,
		defaultChainId,
	)

	result, errMap := service.ProcessBlock(s, block, false)
	assert.Nil(t, errMap)

	ethBlock := result
	assert.Equal(t, "0x7b", *ethBlock.Number) // 123 in hex
	assert.Equal(t, "0x123abc", *ethBlock.Hash)
	assert.Equal(t, "0x456def", ethBlock.ParentHash)
	assert.Equal(t, "0x3e8", ethBlock.GasUsed)     // 1000 in hex
	assert.Equal(t, "0x7d0", ethBlock.Size)        // 2000 in hex
	assert.Equal(t, 2, len(ethBlock.Transactions)) // Only SUCCESS transactions
}

func TestProcessBlock_WithLongHashes(t *testing.T) {
	ctrl, mockClient, logger := setupTest(t)
	defer ctrl.Finish()

	longHash := "0x123abc" + strings.Repeat("0", 100) // Hash longer than 66 chars
	block := &domain.BlockResponse{
		Number:       123,
		Hash:         longHash,
		PreviousHash: longHash,
		GasUsed:      1000,
		Size:         2000,
		LogsBloom:    "0x0",
		Timestamp: domain.Timestamp{
			From: "1640995200",
		},
	}

	mockClient.EXPECT().
		GetContractResults(block.Timestamp).
		Return([]domain.ContractResults{})

	s := service.NewEthService(
		nil,
		mockClient,
		logger,
		nil,
		defaultChainId,
	)

	result, errMap := service.ProcessBlock(s, block, false)
	assert.Nil(t, errMap)

	ethBlock := result
	assert.Equal(t, 66, len(*ethBlock.Hash))
	assert.Equal(t, 66, len(ethBlock.ParentHash))
}

func TestProcessTransaction_LegacyTransaction(t *testing.T) {
	// Create a properly formatted Ethereum address by padding with zeros
	toAddress := "0xto123" + strings.Repeat("0", 35) // 0x + 40 hex chars = 42 total length

	contractResult := domain.ContractResults{
		BlockNumber:        123,
		BlockHash:          "0xblockHash123",
		Hash:               "0xtxHash123" + strings.Repeat("0", 100),
		From:               "0xfrom123" + strings.Repeat("0", 100),
		To:                 toAddress,
		GasUsed:            1000,
		TransactionIndex:   5,
		Amount:             100,
		V:                  27,
		R:                  "0xr123" + strings.Repeat("0", 100),
		S:                  "0xs123" + strings.Repeat("0", 100),
		Nonce:              10,
		Type:               0, // Legacy transaction
		GasPrice:           "0x100",
		FunctionParameters: "0xabcd",
		ChainID:            "0x1",
	}

	result := service.ProcessTransaction(contractResult)
	tx, ok := result.(domain.Transaction)
	assert.True(t, ok)

	// Verify field conversions
	assert.Equal(t, "0x7b", *tx.BlockNumber) // 123 in hex
	assert.Equal(t, "0xblockHash123", *tx.BlockHash)
	assert.Equal(t, "0xtxHash123", tx.Hash[:11]) // Verify hash is trimmed
	assert.Equal(t, "0xfrom123", tx.From[:9])    // Verify from address is trimmed
	assert.Equal(t, 42, len(*tx.To))             // Verify to address is exactly 42 chars (standard ETH address length)
	assert.Equal(t, toAddress, *tx.To)           // Verify complete to address
	assert.Equal(t, "0x3e8", tx.Gas)             // 1000 in hex
	assert.Equal(t, "0x5", *tx.TransactionIndex)
	assert.Equal(t, "0x64", tx.Value) // 100 in hex
	assert.Equal(t, "0x1b", tx.V)     // 27 in hex
	assert.Equal(t, 66, len(tx.R))    // Verify R length
	assert.Equal(t, 66, len(tx.S))    // Verify S length
	assert.Equal(t, "0xa", tx.Nonce)  // 10 in hex
	assert.Equal(t, "0x0", tx.Type)   // Type 0
	assert.Equal(t, "0x100", tx.GasPrice)
	assert.Equal(t, "0xabcd", tx.Input)
	assert.Equal(t, "0x1", *tx.ChainId)
}

func TestProcessTransaction_EIP2930(t *testing.T) {
	toAddress := "0xto123" + strings.Repeat("0", 35) // Properly formatted Ethereum address

	contractResult := domain.ContractResults{
		BlockNumber: 123,
		Hash:        "0xtxHash123" + strings.Repeat("0", 100),
		From:        "0xfrom123" + strings.Repeat("0", 100),
		To:          toAddress,
		Type:        1, // EIP-2930
		GasPrice:    "0x100",
		R:           "0xr123" + strings.Repeat("0", 100),
		S:           "0xs123" + strings.Repeat("0", 100),
	}

	result := service.ProcessTransaction(contractResult)
	tx, ok := result.(domain.Transaction2930)
	assert.True(t, ok)
	assert.Empty(t, tx.AccessList)
	assert.Equal(t, "0x1", tx.Type)
	assert.Equal(t, toAddress, *tx.Transaction.To)
}

func TestProcessTransaction_EIP1559(t *testing.T) {
	toAddress := "0xto123" + strings.Repeat("0", 35) // Properly formatted Ethereum address

	contractResult := domain.ContractResults{
		BlockNumber:          123,
		Hash:                 "0xtxHash123" + strings.Repeat("0", 100),
		From:                 "0xfrom123" + strings.Repeat("0", 100),
		To:                   toAddress,
		Type:                 2, // EIP-1559
		MaxPriorityFeePerGas: "0x100",
		MaxFeePerGas:         "0x200",
		R:                    "0xr123" + strings.Repeat("0", 100),
		S:                    "0xs123" + strings.Repeat("0", 100),
	}

	result := service.ProcessTransaction(contractResult)
	tx, ok := result.(domain.Transaction1559)
	assert.True(t, ok)
	assert.Empty(t, tx.AccessList)
	assert.Equal(t, "0x2", tx.Type)
	assert.Equal(t, "0x100", tx.MaxPriorityFeePerGas)
	assert.Equal(t, "0x200", tx.MaxFeePerGas)
	assert.Equal(t, toAddress, *tx.Transaction.To)
}

func TestProcessTransaction_UnknownType(t *testing.T) {
	toAddress := "0xto123" + strings.Repeat("0", 35) // Properly formatted Ethereum address

	contractResult := domain.ContractResults{
		BlockNumber: 123,
		Hash:        "0xtxHash123" + strings.Repeat("0", 100),
		From:        "0xfrom123" + strings.Repeat("0", 100),
		To:          toAddress,
		Type:        99, // Unknown type
		R:           "0xr123" + strings.Repeat("0", 100),
		S:           "0xs123" + strings.Repeat("0", 100),
	}

	result := service.ProcessTransaction(contractResult)
	tx, ok := result.(domain.Transaction)
	assert.True(t, ok)
	assert.Equal(t, "0x63", tx.Type) // 99 in hex
	assert.Equal(t, toAddress, *tx.To)
}

func TestFormatTransactionCallObject(t *testing.T) {
	ctrl, _, logger := setupTest(t)
	defer ctrl.Finish()

	s := service.NewEthService(
		nil,
		nil,
		logger,
		nil,
		defaultChainId,
	)

	testCases := []struct {
		name        string
		input       *domain.TransactionCallObject
		blockParam  interface{}
		estimate    bool
		expected    map[string]interface{}
		expectError bool
	}{
		{
			name: "Basic transaction with value",
			input: &domain.TransactionCallObject{
				From:  "0x123",
				To:    "0x456",
				Value: "0x64", // 100 in hex
			},
			blockParam: nil,
			estimate:   false,
			expected: map[string]interface{}{
				"from":     "0x123",
				"to":       "0x456",
				"value":    "0", // 100 weibars is less than 1 tinybar, so it rounds to 0
				"estimate": false,
			},
			expectError: false,
		},
		{
			name: "Transaction with gas price",
			input: &domain.TransactionCallObject{
				GasPrice: "0x64", // 100 in hex
			},
			blockParam: nil,
			estimate:   true,
			expected: map[string]interface{}{
				"gasPrice": "100",
				"estimate": true,
			},
			expectError: false,
		},
		{
			name: "Transaction with gas",
			input: &domain.TransactionCallObject{
				Gas: "0x64", // 100 in hex
			},
			blockParam: "latest",
			estimate:   false,
			expected: map[string]interface{}{
				"gas":      "100",
				"block":    "latest",
				"estimate": false,
			},
			expectError: false,
		},
		{
			name: "Transaction with input and data",
			input: &domain.TransactionCallObject{
				Input: "0x123",
				Data:  "0x123",
			},
			blockParam: nil,
			estimate:   false,
			expected: map[string]interface{}{
				"data":     "0x123",
				"estimate": false,
			},
			expectError: false,
		},
		{
			name: "Error: Conflicting input and data",
			input: &domain.TransactionCallObject{
				Input: "0x123",
				Data:  "0x456",
			},
			blockParam:  nil,
			estimate:    false,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.FormatTransactionCallObject(s, tc.input, tc.blockParam, tc.estimate)
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestWeibarHexToTinyBarInt(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      int64
		expectError   bool
		errorContains string
	}{
		{
			name:     "Zero value",
			input:    "0x0",
			expected: 0,
		},
		{
			name:     "Simple hex value",
			input:    "0x64", // 100 in hex
			expected: 0,      // 100 weibars < 1 tinybar
		},
		{
			name:     "Large hex value",
			input:    "0x2386f26fc10000", // 10000000000000000 in hex
			expected: 1000000,            // 1 million tinybars
		},
		{
			name:     "Decimal string",
			input:    "1000000000000000",
			expected: 100000,
		},
		{
			name:     "Small value rounds up to 1",
			input:    "0x2386f26fc", // Just under 1 tinybar
			expected: 1,
		},
		{
			name:          "Invalid hex string",
			input:         "0xNOTHEX",
			expectError:   true,
			errorContains: "failed to parse hex value",
		},
		{
			name:     "Empty hex string",
			input:    "0x",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.WeibarHexToTinyBarInt(tc.input)
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestNormalizeHexString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Already normalized hex",
			input:    "0x123",
			expected: "0x123",
		},
		{
			name:     "Leading zeros after 0x",
			input:    "0x0000123",
			expected: "0x123",
		},
		{
			name:     "Only zeros",
			input:    "0x0000",
			expected: "0x0",
		},
		{
			name:     "No 0x prefix",
			input:    "123",
			expected: "123",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Just 0x",
			input:    "0x",
			expected: "0x0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.NormalizeHexString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestProcessTransactionResponse(t *testing.T) {
	// Helper function to create a 64-character hex string (without 0x prefix)
	makeHexString := func(char string) string {
		return "0x" + strings.Repeat(char, 64)
	}

	intPtr := func(i int) *int {
		return &i
	}

	testCases := []struct {
		name     string
		input    domain.ContractResultResponse
		expected interface{}
	}{
		{
			name: "Legacy transaction",
			input: domain.ContractResultResponse{
				BlockNumber:      123,
				BlockHash:        makeHexString("1"),
				Hash:             makeHexString("2"),
				From:             "0x" + strings.Repeat("3", 40),
				To:               "0x" + strings.Repeat("4", 40),
				GasUsed:          21000,
				GasPrice:         "0x5678",
				TransactionIndex: 1,
				Amount:           1000000,
				V:                27,
				R:                makeHexString("a"),
				S:                makeHexString("b"),
				Nonce:            5,
				Type:             intPtr(0),
				ChainID:          "0x1",
			},
			expected: domain.Transaction{
				BlockHash:        stringPtr(makeHexString("1")),
				BlockNumber:      stringPtr("0x7b"), // 123 in hex
				From:             "0x" + strings.Repeat("3", 40),
				To:               stringPtr("0x" + strings.Repeat("4", 40)),
				Gas:              "0x5208", // 21000 in hex
				GasPrice:         "0x5678",
				Hash:             makeHexString("2"),
				Nonce:            "0x5",
				TransactionIndex: stringPtr("0x1"),
				Value:            "0xf4240", // 1000000 in hex
				V:                "0x1b",    // 27 in hex
				R:                makeHexString("a"),
				S:                makeHexString("b"),
				Type:             "0x0",
				ChainId:          stringPtr("0x1"),
			},
		},
		{
			name: "EIP-2930 transaction",
			input: domain.ContractResultResponse{
				BlockNumber:      456,
				BlockHash:        makeHexString("5"),
				Hash:             makeHexString("6"),
				From:             "0x" + strings.Repeat("7", 40),
				To:               "0x" + strings.Repeat("8", 40),
				GasUsed:          21000,
				GasPrice:         "0x5678",
				TransactionIndex: 2,
				Amount:           2000000,
				V:                28,
				R:                makeHexString("c"),
				S:                makeHexString("d"),
				Nonce:            6,
				Type:             intPtr(1),
				ChainID:          "0x1",
			},
			expected: domain.Transaction2930{
				Transaction: domain.Transaction{
					BlockHash:        stringPtr(makeHexString("5")),
					BlockNumber:      stringPtr("0x1c8"), // 456 in hex
					From:             "0x" + strings.Repeat("7", 40),
					To:               stringPtr("0x" + strings.Repeat("8", 40)),
					Gas:              "0x5208", // 21000 in hex
					GasPrice:         "0x5678",
					Hash:             makeHexString("6"),
					Nonce:            "0x6",
					TransactionIndex: stringPtr("0x2"),
					Value:            "0x1e8480", // 2000000 in hex
					V:                "0x1c",     // 28 in hex
					R:                makeHexString("c"),
					S:                makeHexString("d"),
					Type:             "0x1",
					ChainId:          stringPtr("0x1"),
				},
				AccessList: []domain.AccessListEntry{},
			},
		},
		{
			name: "EIP-1559 transaction",
			input: domain.ContractResultResponse{
				BlockNumber:          789,
				BlockHash:            makeHexString("9"),
				Hash:                 makeHexString("f"),
				From:                 "0x" + strings.Repeat("a", 40),
				To:                   "0x" + strings.Repeat("b", 40),
				GasUsed:              21000,
				GasPrice:             "0x5678",
				TransactionIndex:     3,
				Amount:               3000000,
				V:                    29,
				R:                    makeHexString("e"),
				S:                    makeHexString("f"),
				Nonce:                7,
				Type:                 intPtr(2),
				ChainID:              "0x1",
				MaxPriorityFeePerGas: "0x1234",
				MaxFeePerGas:         "0x5678",
			},
			expected: domain.Transaction1559{
				Transaction: domain.Transaction{
					BlockHash:        stringPtr(makeHexString("9")),
					BlockNumber:      stringPtr("0x315"), // 789 in hex
					From:             "0x" + strings.Repeat("a", 40),
					To:               stringPtr("0x" + strings.Repeat("b", 40)),
					Gas:              "0x5208", // 21000 in hex
					GasPrice:         "0x5678",
					Hash:             makeHexString("f"),
					Nonce:            "0x7",
					TransactionIndex: stringPtr("0x3"),
					Value:            "0x2dc6c0", // 3000000 in hex
					V:                "0x1d",     // 29 in hex
					R:                makeHexString("e"),
					S:                makeHexString("f"),
					Type:             "0x2",
					ChainId:          stringPtr("0x1"),
				},
				AccessList:           []domain.AccessListEntry{},
				MaxPriorityFeePerGas: "0x1234",
				MaxFeePerGas:         "0x5678",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.ProcessTransactionResponse(tc.input)

			// Type-specific assertions
			switch expected := tc.expected.(type) {
			case domain.Transaction:
				actual, ok := result.(domain.Transaction)
				assert.True(t, ok, "Expected Transaction type")
				assert.Equal(t, expected, actual)

			case domain.Transaction2930:
				actual, ok := result.(domain.Transaction2930)
				assert.True(t, ok, "Expected Transaction2930 type")
				assert.Equal(t, expected, actual)

			case domain.Transaction1559:
				actual, ok := result.(domain.Transaction1559)
				assert.True(t, ok, "Expected Transaction1559 type")
				assert.Equal(t, expected, actual)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
