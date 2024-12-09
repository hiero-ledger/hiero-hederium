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

	ethBlock, ok := result.(*domain.Block)
	assert.True(t, ok)
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

	ethBlock, ok := result.(*domain.Block)
	assert.True(t, ok)
	assert.Equal(t, 66, len(*ethBlock.Hash))
	assert.Equal(t, 66, len(ethBlock.ParentHash))
}
