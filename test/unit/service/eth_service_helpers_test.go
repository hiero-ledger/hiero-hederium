package service_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/georgi-l95/Hederium/internal/domain"
	"github.com/georgi-l95/Hederium/internal/service"
	"github.com/georgi-l95/Hederium/test/unit/mocks"
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

	result, errMap := service.GetFeeWeibars(s)
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

	result, errMap := service.GetFeeWeibars(s)
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
