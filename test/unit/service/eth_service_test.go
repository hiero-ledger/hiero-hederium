package service_test

import (
	"fmt"
	"strconv"
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
		GetNetworkFees("", "").
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
		GetNetworkFees("", "").
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
					Return(map[string]interface{}{
						"number": float64(123),
					}, nil)
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

func TestGetBalance_Latest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create mock client
	mockClient := mocks.NewMockMirrorClient(ctrl)

	// Setup expectations for getting latest block
	mockClient.EXPECT().
		GetLatestBlock().
		Return(map[string]interface{}{"number": float64(42)}, nil)

	// Setup expectations for getting block by number
	mockClient.EXPECT().
		GetBlockByHashOrNumber("42").
		Return(&domain.BlockResponse{
			Timestamp: domain.Timestamp{
				To: "1234567890.000000000",
			},
		})

	// Setup expectations for getting balance
	mockClient.EXPECT().
		GetBalance("0x123", "1234567890.000000000").
		Return("0x2a")

	s := service.NewEthService(
		nil, // hClient not needed for this test
		mockClient,
		logger,
		nil, // tieredLimiter not needed for this test
		defaultChainId,
	)

	result := s.GetBalance("0x123", "latest")
	assert.Equal(t, "0x2a", result)
}

func TestGetBalance_Earliest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	mockClient := mocks.NewMockMirrorClient(ctrl)

	// Setup expectations for getting block zero
	mockClient.EXPECT().
		GetBlockByHashOrNumber("0").
		Return(&domain.BlockResponse{
			Timestamp: domain.Timestamp{
				To: "0.000000000",
			},
		})

	// Setup expectations for getting balance
	mockClient.EXPECT().
		GetBalance("0x123", "0.000000000").
		Return("0x0")

	s := service.NewEthService(
		nil,
		mockClient,
		logger,
		nil,
		defaultChainId,
	)

	result := s.GetBalance("0x123", "earliest")
	assert.Equal(t, "0x0", result)
}

func TestGetBalance_SpecificBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	mockClient := mocks.NewMockMirrorClient(ctrl)

	// Setup expectations for getting specific block
	mockClient.EXPECT().
		GetBlockByHashOrNumber("100").
		Return(&domain.BlockResponse{
			Timestamp: domain.Timestamp{
				To: "1234567890.000000000",
			},
		})

	// Setup expectations for getting balance
	mockClient.EXPECT().
		GetBalance("0x123", "1234567890.000000000").
		Return("0x64")

	s := service.NewEthService(
		nil,
		mockClient,
		logger,
		nil,
		defaultChainId,
	)

	result := s.GetBalance("0x123", "0x64") // hex for 100
	assert.Equal(t, "0x64", result)
}

func TestGetBalance_BlockNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	mockClient := mocks.NewMockMirrorClient(ctrl)

	// Setup expectations for getting block that doesn't exist
	mockClient.EXPECT().
		GetBlockByHashOrNumber("999999").
		Return(nil)

	s := service.NewEthService(
		nil,
		mockClient,
		logger,
		nil,
		defaultChainId,
	)

	result := s.GetBalance("0x123", "999999")
	assert.Equal(t, "0x0", result)
}

func TestCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	testCases := []struct {
		name           string
		transaction    interface{}
		blockParam     interface{}
		mockResponse   interface{}
		expectedResult interface{}
		expectError    bool
		setupMock      bool
	}{
		{
			name: "Successful call",
			transaction: map[string]interface{}{
				"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				"data": "0x70a08231000000000000000000000000b1d6b01b94d854f521665696ea17fcf87c160d97",
			},
			blockParam:     "latest",
			mockResponse:   "0x0000000000000000000000000000000000000000000000000000000000000064",
			expectedResult: "0x64",
			expectError:    false,
			setupMock:      true,
		},
		{
			name: "Invalid transaction object",
			transaction: map[string]interface{}{
				"input": "0x123",
				"data":  "0x456", // Conflicting input and data
			},
			blockParam:     "latest",
			mockResponse:   nil,
			expectedResult: "0x0",
			expectError:    true,
			setupMock:      false,
		},
		{
			name: "Empty response from mirror node",
			transaction: map[string]interface{}{
				"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				"data": "0x70a08231",
			},
			blockParam:     "latest",
			mockResponse:   nil,
			expectedResult: "0x0",
			expectError:    true,
			setupMock:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMock {
				mockClient.EXPECT().
					PostCall(gomock.Any()).
					Return(tc.mockResponse).
					Times(1)
			}

			result, errMap := s.Call(tc.transaction, tc.blockParam)

			if tc.expectError {
				assert.NotNil(t, errMap)
				assert.Equal(t, -32000, errMap["code"])
			} else {
				assert.Nil(t, errMap)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestEstimateGas(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	testCases := []struct {
		name           string
		transaction    interface{}
		blockParam     interface{}
		mockResponse   interface{}
		expectedResult string
		expectError    bool
		setupMock      bool
	}{
		{
			name: "Successful gas estimation",
			transaction: map[string]interface{}{
				"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				"data": "0x70a08231000000000000000000000000b1d6b01b94d854f521665696ea17fcf87c160d97",
			},
			blockParam:     "latest",
			mockResponse:   "0x5208", // 21000 gas
			expectedResult: "0x5208",
			expectError:    false,
			setupMock:      true,
		},
		{
			name: "Invalid transaction object",
			transaction: map[string]interface{}{
				"input": "0x123",
				"data":  "0x456", // Conflicting input and data
			},
			blockParam:     "latest",
			mockResponse:   nil,
			expectedResult: "0x0",
			expectError:    true,
			setupMock:      false,
		},
		{
			name: "Empty response from mirror node",
			transaction: map[string]interface{}{
				"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				"data": "0x70a08231",
			},
			blockParam:     "latest",
			mockResponse:   nil,
			expectedResult: "0x0",
			expectError:    true,
			setupMock:      true,
		},
		{
			name: "Zero gas estimation",
			transaction: map[string]interface{}{
				"to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			},
			blockParam:     "latest",
			mockResponse:   "0x0",
			expectedResult: "0x0",
			expectError:    false,
			setupMock:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMock {
				mockClient.EXPECT().
					PostCall(gomock.Any()).
					Return(tc.mockResponse).
					Times(1)
			}

			result, errMap := s.EstimateGas(tc.transaction, tc.blockParam)

			if tc.expectError {
				assert.NotNil(t, errMap)
				assert.Equal(t, -32000, errMap["code"])
			} else {
				assert.Nil(t, errMap)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetTransactionByHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	// Common test data
	testHash := "0x5d019848d6dad96bc3a9e947350975cd16cf1c51efd4d5b9a273803446fbbb43"
	baseContractResult := domain.ContractResultResponse{
		BlockNumber:        123,
		BlockHash:          "0x" + strings.Repeat("1", 64),
		Hash:               testHash,
		From:               "0x" + strings.Repeat("2", 40),
		To:                 "0x" + strings.Repeat("3", 40),
		GasUsed:            21000,
		GasPrice:           "0x5678",
		TransactionIndex:   1,
		Amount:             1000000,
		V:                  27,
		R:                  "0x" + strings.Repeat("4", 64),
		S:                  "0x" + strings.Repeat("5", 64),
		Nonce:              5,
		FunctionParameters: "0x",
		ChainID:            defaultChainId,
		Type:               new(int),
	}

	testCases := []struct {
		name           string
		hash           string
		mockResult     interface{}
		expectedResult bool
		checkFields    func(t *testing.T, result interface{})
	}{
		{
			name: "Legacy transaction (type 0)",
			hash: testHash,
			mockResult: func() domain.ContractResultResponse {
				result := baseContractResult
				typeVal := 0
				result.Type = &typeVal
				return result
			}(),
			expectedResult: true,
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction)
				assert.True(t, ok)
				assert.Equal(t, "0x0", tx.Type)
				assert.Equal(t, testHash, tx.Hash)
				assert.Equal(t, "0x7b", *tx.BlockNumber) // 123 in hex
				assert.Equal(t, defaultChainId, *tx.ChainId)
			},
		},
		{
			name: "EIP-2930 transaction (type 1)",
			hash: testHash,
			mockResult: func() domain.ContractResultResponse {
				result := baseContractResult
				typeVal := 1
				result.Type = &typeVal
				return result
			}(),
			expectedResult: true,
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction2930)
				assert.True(t, ok)
				assert.Equal(t, "0x1", tx.Type)
				assert.Empty(t, tx.AccessList)
				assert.Equal(t, testHash, tx.Hash)
				assert.Equal(t, "0x7b", *tx.BlockNumber) // 123 in hex
				assert.Equal(t, defaultChainId, *tx.ChainId)
			},
		},
		{
			name: "EIP-1559 transaction (type 2)",
			hash: testHash,
			mockResult: func() domain.ContractResultResponse {
				result := baseContractResult
				typeVal := 2
				result.Type = &typeVal
				result.MaxPriorityFeePerGas = "0x1234"
				result.MaxFeePerGas = "0x5678"
				return result
			}(),
			expectedResult: true,
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction1559)
				assert.True(t, ok)
				assert.Equal(t, "0x2", tx.Type)
				assert.Empty(t, tx.AccessList)
				assert.Equal(t, "0x1234", tx.MaxPriorityFeePerGas)
				assert.Equal(t, "0x5678", tx.MaxFeePerGas)
				assert.Equal(t, testHash, tx.Hash)
			},
		},
		{
			name:           "Transaction not found",
			hash:           testHash,
			mockResult:     nil,
			expectedResult: false,
			checkFields:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.EXPECT().
				GetContractResult(tc.hash).
				Return(tc.mockResult).
				Times(1)

			result := s.GetTransactionByHash(tc.hash)
			if tc.checkFields != nil {
				tc.checkFields(t, result)
			}
		})
	}
}

func TestGetTransactionReceipt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()

	t.Run("Success case with full transaction receipt", func(t *testing.T) {
		mockClient := mocks.NewMockMirrorClient(ctrl)
		s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

		txHash := "0x123456"
		blockHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		blockNumber := int64(12345)
		transactionIndex := 1
		gasUsed := int64(21000)
		blockGasUsed := int64(100000)
		from := "0xabc"
		to := "0xdef"
		contractAddress := "0xcontract"
		gasPrice := "5000"
		txType := 2

		// Mock contract result with logs
		contractResult := domain.ContractResultResponse{
			BlockHash:        blockHash + "extra", // Adding extra to test trimming to 66 chars
			BlockNumber:      blockNumber,
			TransactionIndex: transactionIndex,
			GasUsed:          gasUsed,
			BlockGasUsed:     blockGasUsed,
			From:             from,
			To:               to,
			Address:          contractAddress,
			GasPrice:         gasPrice,
			Status:           "0x1",
			Type:             &txType,
			Logs: []domain.MirroNodeLogs{
				{
					Address: "0xlog1",
					Data:    "0xdata1",
					Topics:  []string{"0xtopic1", "0xtopic2"},
				},
				{
					Address: "0xlog2",
					Data:    "0xdata2",
					Topics:  []string{"0xtopic3"},
				},
			},
			Bloom: "0x1234", // Custom bloom value
		}

		mockClient.EXPECT().
			GetContractResult(txHash).
			Return(contractResult)

		result := s.GetTransactionReceipt(txHash)
		receipt, ok := result.(domain.TransactionReceipt)
		assert.True(t, ok, "Result should be of type domain.TransactionReceipt")

		// Verify all fields
		assert.Equal(t, blockHash[:66], receipt.BlockHash)
		assert.Equal(t, "0x"+strconv.FormatInt(blockNumber, 16), receipt.BlockNumber)
		assert.Equal(t, from, receipt.From)
		assert.Equal(t, to, receipt.To)
		assert.Equal(t, "0x"+strconv.FormatInt(blockGasUsed, 16), receipt.CumulativeGasUsed)
		assert.Equal(t, "0x"+strconv.FormatInt(gasUsed, 16), receipt.GasUsed)
		assert.Equal(t, contractAddress, receipt.ContractAddress)
		assert.Equal(t, txHash, receipt.TransactionHash)
		assert.Equal(t, "0x"+strconv.FormatInt(int64(transactionIndex), 16), receipt.TransactionIndex)
		assert.Equal(t, "0x"+gasPrice, receipt.EffectiveGasPrice)
		assert.Equal(t, "0x1", receipt.Status)
		assert.Equal(t, "0x"+strconv.FormatInt(int64(txType), 16), *receipt.Type)
		assert.Equal(t, "0x1234", receipt.LogsBloom)

		// Verify logs
		assert.Len(t, receipt.Logs, 2)
		assert.Equal(t, "0xlog1", receipt.Logs[0].Address)
		assert.Equal(t, "0xdata1", receipt.Logs[0].Data)
		assert.Equal(t, []string{"0xtopic1", "0xtopic2"}, receipt.Logs[0].Topics)
		assert.Equal(t, "0x0", receipt.Logs[0].LogIndex)
		assert.Equal(t, txHash, receipt.Logs[0].TransactionHash)
		assert.Equal(t, blockHash[:66], receipt.Logs[0].BlockHash)
		assert.Equal(t, receipt.BlockNumber, receipt.Logs[0].BlockNumber)
		assert.Equal(t, receipt.TransactionIndex, receipt.Logs[0].TransactionIndex)
		assert.False(t, receipt.Logs[0].Removed)
	})

	t.Run("Transaction not found", func(t *testing.T) {
		mockClient := mocks.NewMockMirrorClient(ctrl)
		s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

		txHash := "0xnonexistent"
		mockClient.EXPECT().
			GetContractResult(txHash).
			Return(nil)

		result := s.GetTransactionReceipt(txHash)
		assert.Nil(t, result)
	})

	t.Run("Empty bloom filter", func(t *testing.T) {
		mockClient := mocks.NewMockMirrorClient(ctrl)
		s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

		txHash := "0x123456"
		contractResult := domain.ContractResultResponse{
			BlockHash:   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			BlockNumber: 12345,
			Status:      "0x1",
			Bloom:       "0x", // Empty bloom
		}

		mockClient.EXPECT().
			GetContractResult(txHash).
			Return(contractResult)

		result := s.GetTransactionReceipt(txHash)
		receipt, ok := result.(domain.TransactionReceipt)
		assert.True(t, ok)
		assert.Equal(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", receipt.LogsBloom)
	})

	t.Run("Nil transaction type", func(t *testing.T) {
		mockClient := mocks.NewMockMirrorClient(ctrl)
		s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

		txHash := "0x123456"
		contractResult := domain.ContractResultResponse{
			BlockHash:   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			BlockNumber: 12345,
			Status:      "0x1",
			Type:        nil, // Nil type
		}

		mockClient.EXPECT().
			GetContractResult(txHash).
			Return(contractResult)

		result := s.GetTransactionReceipt(txHash)
		receipt, ok := result.(domain.TransactionReceipt)
		assert.True(t, ok)
		assert.Nil(t, receipt.Type)
	})
}

func TestFeeHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	testCases := []struct {
		name              string
		blockCount        string
		newestBlock       string
		rewardPercentiles []string
		mockLatestBlock   map[string]interface{}
		expectNil         bool
		expectError       bool
		setupMocks        func()
		validateResult    func(t *testing.T, result interface{})
	}{
		{
			name:              "Success with fixed fee",
			blockCount:        "0x5",
			newestBlock:       "0x64", // hex for 100
			rewardPercentiles: []string{"0x5", "0xa", "0xf"},
			mockLatestBlock: map[string]interface{}{
				"number": float64(100),
			},
			expectNil:   false,
			expectError: false,
			setupMocks: func() {
				// Mock GetLatestBlock for initial block number check
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": float64(100),
					}, nil).Times(2)

				// Mock GetBlockByHashOrNumber for each block in the range
				for i := 96; i <= 100; i++ {
					mockClient.EXPECT().
						GetBlockByHashOrNumber(strconv.Itoa(i)).
						Return(&domain.BlockResponse{
							Number:  i,
							GasUsed: 50000,
						})
				}

				// Mock GetGasPrice via GetNetworkFees for each block
				mockClient.EXPECT().
					GetNetworkFees("", "desc").
					Return(int64(10000000), nil).Times(5)
			},
			validateResult: func(t *testing.T, result interface{}) {
				feeHistory, ok := result.(*domain.FeeHistory)
				assert.True(t, ok)
				assert.Equal(t, "0x60", feeHistory.OldestBlock)   // 96 in hex (100-5+1)
				assert.Equal(t, 6, len(feeHistory.BaseFeePerGas)) // blockCount + 1
				assert.Equal(t, 5, len(feeHistory.GasUsedRatio))
				assert.Equal(t, 5, len(feeHistory.Reward))
			},
		},
		{
			name:              "Invalid block count",
			blockCount:        "invalid",
			newestBlock:       "latest",
			rewardPercentiles: []string{},
			mockLatestBlock: map[string]interface{}{
				"number": float64(100),
			},
			expectNil:   true,
			expectError: true,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": float64(100),
					}, nil).Times(2)
			},
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name:              "GetLatestBlock error",
			blockCount:        "0x5",
			newestBlock:       "latest",
			rewardPercentiles: []string{},
			expectNil:         true,
			expectError:       true,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(nil, fmt.Errorf("error getting latest block"))
			},
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name:              "Block count exceeds maximum",
			blockCount:        "0x14", // 20 blocks, more than max (10)
			newestBlock:       "latest",
			rewardPercentiles: []string{},
			mockLatestBlock: map[string]interface{}{
				"number": float64(100),
			},
			expectNil:   false,
			expectError: false,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": float64(100),
					}, nil).Times(2)

				// Mock GetBlockByHashOrNumber for each block in the range
				for i := 91; i <= 100; i++ {
					mockClient.EXPECT().
						GetBlockByHashOrNumber(strconv.Itoa(i)).
						Return(&domain.BlockResponse{
							Number:  i,
							GasUsed: 50000,
						})
				}

				// Mock GetGasPrice via GetNetworkFees for each block
				mockClient.EXPECT().
					GetNetworkFees("", "desc").
					Return(int64(100000), nil).Times(10)
			},
			validateResult: func(t *testing.T, result interface{}) {
				feeHistory, ok := result.(*domain.FeeHistory)
				assert.True(t, ok)
				assert.Equal(t, "0x5b", feeHistory.OldestBlock)    // 91 in hex (100-10+1)
				assert.Equal(t, 11, len(feeHistory.BaseFeePerGas)) // maxBlockCountForResult + 1
				assert.Equal(t, 10, len(feeHistory.GasUsedRatio))  // maxBlockCountForResult
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)
			result, errMap := s.FeeHistory(tc.blockCount, tc.newestBlock, tc.rewardPercentiles)

			if tc.expectError {
				assert.NotNil(t, errMap)
				assert.Equal(t, -32000, errMap["code"])
			} else {
				assert.Nil(t, errMap)
			}

			tc.validateResult(t, result)
		})
	}
}

func TestGetStorageAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId)

	testCases := []struct {
		name           string
		address        string
		slot           string
		blockParam     string
		mockBlock      *domain.BlockResponse
		mockState      *domain.ContractStateResponse
		expectedResult interface{}
		expectError    bool
		setupMock      func()
	}{
		{
			name:       "Success with latest block",
			address:    "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			slot:       "0x0",
			blockParam: "latest",
			mockBlock: &domain.BlockResponse{
				Timestamp: domain.Timestamp{
					To: "2023-12-09T12:00:00.000Z",
				},
			},
			mockState: &domain.ContractStateResponse{
				State: []domain.ContractState{
					{
						Value: "0x0000000000000000000000000000000000000000000000000000000000000064",
					},
				},
			},
			expectedResult: "0x0000000000000000000000000000000000000000000000000000000000000064",
			expectError:    false,
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
					GetContractStateByAddressAndSlot(
						"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"0x0",
						"2023-12-09T12:00:00.000Z",
					).
					Return(&domain.ContractStateResponse{
						State: []domain.ContractState{
							{
								Value: "0x0000000000000000000000000000000000000000000000000000000000000064",
							},
						},
					}, nil)
			},
		},
		{
			name:       "Success with earliest block",
			address:    "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			slot:       "0x1",
			blockParam: "earliest",
			mockBlock: &domain.BlockResponse{
				Timestamp: domain.Timestamp{
					To: "2023-01-01T00:00:00.000Z",
				},
			},
			mockState: &domain.ContractStateResponse{
				State: []domain.ContractState{
					{
						Value: "0x0000000000000000000000000000000000000000000000000000000000000032",
					},
				},
			},
			expectedResult: "0x0000000000000000000000000000000000000000000000000000000000000032",
			expectError:    false,
			setupMock: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							To: "2023-01-01T00:00:00.000Z",
						},
					})
				mockClient.EXPECT().
					GetContractStateByAddressAndSlot(
						"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"0x1",
						"2023-01-01T00:00:00.000Z",
					).
					Return(&domain.ContractStateResponse{
						State: []domain.ContractState{
							{
								Value: "0x0000000000000000000000000000000000000000000000000000000000000032",
							},
						},
					}, nil)
			},
		},
		{
			name:       "Success with specific block number",
			address:    "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			slot:       "0x2",
			blockParam: "0x50",
			mockBlock: &domain.BlockResponse{
				Timestamp: domain.Timestamp{
					To: "2023-06-01T00:00:00.000Z",
				},
			},
			mockState: &domain.ContractStateResponse{
				State: []domain.ContractState{
					{
						Value: "0x0000000000000000000000000000000000000000000000000000000000000096",
					},
				},
			},
			expectedResult: "0x0000000000000000000000000000000000000000000000000000000000000096",
			expectError:    false,
			setupMock: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("80").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							To: "2023-06-01T00:00:00.000Z",
						},
					})
				mockClient.EXPECT().
					GetContractStateByAddressAndSlot(
						"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"0x2",
						"2023-06-01T00:00:00.000Z",
					).
					Return(&domain.ContractStateResponse{
						State: []domain.ContractState{
							{
								Value: "0x0000000000000000000000000000000000000000000000000000000000000096",
							},
						},
					}, nil)
			},
		},
		{
			name:        "Block not found",
			address:     "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			slot:        "0x0",
			blockParam:  "0x999",
			mockBlock:   nil,
			mockState:   nil,
			expectError: true,
			setupMock: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("2457").
					Return(nil)
			},
		},
		{
			name:       "Empty state response",
			address:    "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			slot:       "0x3",
			blockParam: "latest",
			mockBlock: &domain.BlockResponse{
				Timestamp: domain.Timestamp{
					To: "2023-12-09T12:00:00.000Z",
				},
			},
			mockState:      &domain.ContractStateResponse{State: []domain.ContractState{}},
			expectedResult: "0x0000000000000000000000000000000000000000000000000000000000000000",
			expectError:    false,
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
					GetContractStateByAddressAndSlot(
						"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"0x3",
						"2023-12-09T12:00:00.000Z",
					).
					Return(&domain.ContractStateResponse{State: []domain.ContractState{}}, nil)
			},
		},
		{
			name:       "Error getting state",
			address:    "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			slot:       "0x4",
			blockParam: "latest",
			mockBlock: &domain.BlockResponse{
				Timestamp: domain.Timestamp{
					To: "2023-12-09T12:00:00.000Z",
				},
			},
			mockState:   nil,
			expectError: true,
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
					GetContractStateByAddressAndSlot(
						"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"0x4",
						"2023-12-09T12:00:00.000Z",
					).
					Return(nil, fmt.Errorf("failed to get storage"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, errMap := s.GetStorageAt(tc.address, tc.slot, tc.blockParam)

			if tc.expectError {
				assert.NotNil(t, errMap)
				assert.Equal(t, -32000, errMap["code"])
			} else {
				assert.Nil(t, errMap)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
