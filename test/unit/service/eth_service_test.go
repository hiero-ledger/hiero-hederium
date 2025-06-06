package service_test

import (
	"fmt"
	"testing"

	"github.com/georgi-l95/Hederium/internal/domain"
	"github.com/georgi-l95/Hederium/internal/service"
	"github.com/georgi-l95/Hederium/test/unit/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultChainId = "0x127" // Default chain ID for tests

func TestGetBlockNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create mock client from the interface
	mockClient := mocks.NewMockMirrorClient(ctrl)
	mockClient.EXPECT().
		GetLatestBlock().
		Return(map[string]interface{}{"number": float64(42)}, nil)

	s := service.NewEthService(
		nil,        // hClient not needed for this test
		mockClient, // pass the mock as the interface
		logger,
		nil, // tieredLimiter not needed for this test
		defaultChainId,
	)

	result, errMap := s.GetBlockNumber()
	assert.Nil(t, errMap)
	// 42 in hex is "0x2a"
	assert.Equal(t, "0x2a", result)
}

func TestGetAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create mock client
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(
		nil,        // hClient not needed for this test
		mockClient, // pass the mock as the interface
		logger,
		nil, // tieredLimiter not needed for this test
		defaultChainId,
	)

	result, errMap := s.GetAccounts()
	assert.Nil(t, errMap)

	accounts, ok := result.([]string)
	assert.True(t, ok, "Result should be of type []string")
	assert.Empty(t, accounts, "Accounts array should be empty")
}

func TestSyncing(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	// Test
	result, errMap := s.Syncing()
	assert.Nil(t, errMap)
	assert.Equal(t, false, result)
}

func TestMining(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	// Test
	result, errMap := s.Mining()
	assert.Nil(t, errMap)
	assert.Equal(t, false, result)
}

func TestMaxPriorityFeePerGas(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	// Test
	result, errMap := s.MaxPriorityFeePerGas()
	assert.Nil(t, errMap)
	assert.Equal(t, "0x0", result)
}

func TestHashrate(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	// Test
	result, errMap := s.Hashrate()
	assert.Nil(t, errMap)
	assert.Equal(t, "0x0", result)
}

func TestUncleRelatedMethods(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	// Test all uncle-related methods
	t.Run("GetUncleCountByBlockNumber", func(t *testing.T) {
		result, errMap := s.GetUncleCountByBlockNumber()
		assert.Nil(t, errMap)
		assert.Equal(t, "0x0", result)
	})

	t.Run("GetUncleByBlockNumberAndIndex", func(t *testing.T) {
		result, errMap := s.GetUncleByBlockNumberAndIndex()
		assert.Nil(t, errMap)
		assert.Nil(t, result)
	})

	t.Run("GetUncleCountByBlockHash", func(t *testing.T) {
		result, errMap := s.GetUncleCountByBlockHash()
		assert.Nil(t, errMap)
		assert.Equal(t, "0x0", result)
	})

	t.Run("GetUncleByBlockHashAndIndex", func(t *testing.T) {
		result, errMap := s.GetUncleByBlockHashAndIndex()
		assert.Nil(t, errMap)
		assert.Nil(t, result)
	})
}

func TestGetGasPrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create mock client
	mockClient := mocks.NewMockMirrorClient(ctrl)
	mockClient.EXPECT().
		GetNetworkFees().
		Return(int64(100), nil) // Return 100 tinybars

	s := service.NewEthService(
		nil,        // hClient not needed for this test
		mockClient, // pass the mock as the interface
		logger,
		nil, // tieredLimiter not needed for this test
		defaultChainId,
	)

	result, errMap := s.GetGasPrice()
	assert.Nil(t, errMap)

	// Expected calculation:
	// 100 tinybars * 10^8 (conversion to weibars) = 10000000000
	// Add 10% buffer = 11000000000
	// Convert to hex = 0x28fa6ae00
	assert.Equal(t, "0x28fa6ae00", result)
}

func TestGetGasPrice_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	mockClient.EXPECT().
		GetNetworkFees().
		Return(int64(0), fmt.Errorf("network error"))

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	result, errMap := s.GetGasPrice()
	assert.Nil(t, result)
	assert.NotNil(t, errMap)
	assert.Equal(t, -32000, errMap["code"])
	assert.Equal(t, "Failed to fetch gas price", errMap["message"])
}

func TestGetChainId(t *testing.T) {
	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create mock client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	// Test cases
	testCases := []struct {
		name           string
		chainId        string
		expectedResult interface{}
	}{
		{
			name:           "Mainnet chain ID",
			chainId:        "0x127",
			expectedResult: "0x127",
		},
		{
			name:           "Testnet chain ID",
			chainId:        "0x128",
			expectedResult: "0x128",
		},
		{
			name:           "Empty chain ID",
			chainId:        "",
			expectedResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := service.NewEthService(
				nil,        // hClient not needed for this test
				mockClient, // pass the mock as the interface
				logger,
				nil, // tieredLimiter not needed for this test
				tc.chainId,
			)

			result, errMap := s.GetChainId()
			assert.Nil(t, errMap)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestGetBlockByHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	testHash := "0x123abc"
	expectedBlock := &domain.BlockResponse{
		Number:       123,
		Hash:         testHash,
		PreviousHash: "0x456def",
		GasUsed:      1000,
		Size:         2000,
		LogsBloom:    "0x0",
		Timestamp: domain.Timestamp{
			From: "1640995200",
		},
	}

	// Test cases
	testCases := []struct {
		name         string
		hash         string
		showDetails  bool
		mockResponse *domain.BlockResponse
		mockResults  []domain.ContractResults
		expectNil    bool
	}{
		{
			name:         "Success with transactions",
			hash:         testHash,
			showDetails:  false,
			mockResponse: expectedBlock,
			mockResults: []domain.ContractResults{
				{Hash: "0xtx1", Result: "SUCCESS"},
				{Hash: "0xtx2", Result: "SUCCESS"},
			},
			expectNil: false,
		},
		{
			name:         "Block not found",
			hash:         "0xnonexistent",
			showDetails:  false,
			mockResponse: nil,
			mockResults:  nil,
			expectNil:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.EXPECT().
				GetBlockByHashOrNumber(tc.hash).
				Return(tc.mockResponse)

			if tc.mockResponse != nil {
				mockClient.EXPECT().
					GetContractResults(tc.mockResponse.Timestamp).
					Return(tc.mockResults)
			}

			s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)
			result, errMap := s.GetBlockByHash(tc.hash, tc.showDetails)

			if tc.expectNil {
				assert.Nil(t, result)
				assert.Nil(t, errMap)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, errMap)

				ethBlock, ok := result.(*domain.Block)
				assert.True(t, ok)
				assert.Equal(t, "0x7b", *ethBlock.Number) // 123 in hex
				assert.Equal(t, testHash, *ethBlock.Hash)
				assert.Equal(t, expectedBlock.PreviousHash, ethBlock.ParentHash)
				assert.Equal(t, len(tc.mockResults), len(ethBlock.Transactions))
			}
		})
	}
}

func TestGetBlockByNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	expectedBlock := &domain.BlockResponse{
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

	testCases := []struct {
		name            string
		numberOrTag     string
		showDetails     bool
		mockLatestBlock map[string]interface{}
		mockResponse    *domain.BlockResponse
		mockResults     []domain.ContractResults
		expectNil       bool
		setupMocks      func()
	}{
		{
			name:         "Success with specific number",
			numberOrTag:  "0x7b", // 123 in hex
			showDetails:  false,
			mockResponse: expectedBlock,
			mockResults: []domain.ContractResults{
				{Hash: "0xtx1", Result: "SUCCESS"},
			},
			expectNil: false,
			setupMocks: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("123").
					Return(expectedBlock)
			},
		},
		{
			name:            "Success with latest tag",
			numberOrTag:     "latest",
			showDetails:     false,
			mockLatestBlock: map[string]interface{}{"number": float64(123)},
			mockResponse:    expectedBlock,
			mockResults: []domain.ContractResults{
				{Hash: "0xtx1", Result: "SUCCESS"},
			},
			expectNil: false,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(123)}, nil)
				mockClient.EXPECT().
					GetBlockByHashOrNumber("123").
					Return(expectedBlock)
			},
		},
		{
			name:         "Success with earliest tag",
			numberOrTag:  "earliest",
			showDetails:  false,
			mockResponse: expectedBlock,
			mockResults: []domain.ContractResults{
				{Hash: "0xtx1", Result: "SUCCESS"},
			},
			expectNil: false,
			setupMocks: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0").
					Return(expectedBlock)
			},
		},
		{
			name:         "Block not found",
			numberOrTag:  "0x999",
			showDetails:  false,
			mockResponse: nil,
			expectNil:    true,
			setupMocks: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("2457"). // 0x999 in decimal
					Return(nil)
			},
		},
		{
			name:        "Invalid hex number",
			numberOrTag: "0xinvalid",
			showDetails: false,
			expectNil:   true,
			setupMocks:  func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			if tc.mockResponse != nil {
				mockClient.EXPECT().
					GetContractResults(tc.mockResponse.Timestamp).
					Return(tc.mockResults)
			}

			s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)
			result, errMap := s.GetBlockByNumber(tc.numberOrTag, tc.showDetails)

			if tc.expectNil {
				assert.Nil(t, result)
				assert.Nil(t, errMap)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, errMap)

				ethBlock, ok := result.(*domain.Block)
				assert.True(t, ok)
				assert.Equal(t, "0x7b", *ethBlock.Number) // 123 in hex
				assert.Equal(t, expectedBlock.Hash, *ethBlock.Hash)
				assert.Equal(t, expectedBlock.PreviousHash, ethBlock.ParentHash)
				assert.Equal(t, len(tc.mockResults), len(ethBlock.Transactions))
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	testCases := []struct {
		name           string
		address        string
		blockParam     string
		setupMock      func()
		expectedResult string
	}{
		{
			name:       "Latest block balance",
			address:    "0x1234567890123456789012345678901234567890",
			blockParam: "latest",
			setupMock: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)
				mockClient.EXPECT().
					GetBlockByHashOrNumber("100").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							To: "2023-12-09T12:00:00.000Z",
						},
					})
				mockClient.EXPECT().
					GetBalance("0x1234567890123456789012345678901234567890", "2023-12-09T12:00:00.000Z").
					Return("0x64")
			},
			expectedResult: "0x64",
		},
		{
			name:       "Earliest block balance",
			address:    "0x1234567890123456789012345678901234567890",
			blockParam: "earliest",
			setupMock: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							To: "2023-01-01T00:00:00.000Z",
						},
					})
				mockClient.EXPECT().
					GetBalance("0x1234567890123456789012345678901234567890", "2023-01-01T00:00:00.000Z").
					Return("0x32")
			},
			expectedResult: "0x32",
		},
		{
			name:       "Specific block number balance",
			address:    "0x1234567890123456789012345678901234567890",
			blockParam: "0x50",
			setupMock: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("80").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							To: "2023-06-01T00:00:00.000Z",
						},
					})
				mockClient.EXPECT().
					GetBalance("0x1234567890123456789012345678901234567890", "2023-06-01T00:00:00.000Z").
					Return("0x96")
			},
			expectedResult: "0x96",
		},
		{
			name:       "Block not found",
			address:    "0x1234567890123456789012345678901234567890",
			blockParam: "0x999",
			setupMock: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("2457").
					Return(nil)
			},
			expectedResult: "0x0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			result := s.GetBalance(tc.address, tc.blockParam)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
