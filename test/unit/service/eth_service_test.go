package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/service"
	"github.com/LimeChain/Hederium/test/unit/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultChainId = "0x127" // Default chain ID for tests
const GetGasPrice = "eth_gasPrice"
const GetCode = "eth_getCode"
const DefaultExpiration = time.Hour // 1 hour expiration

func TestGetBlockNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create a cache service for testing
	cacheService := mocks.NewMockCacheService(ctrl)

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
		cacheService,
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

	// Create a cache service for testing
	cacheService := mocks.NewMockCacheService(ctrl)

	// Create mock client
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(
		nil,        // hClient not needed for this test
		mockClient, // pass the mock as the interface
		logger,
		nil, // tieredLimiter not needed for this test
		defaultChainId,
		cacheService,
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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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

func TestGetBlockTransactionCountByHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

	testCases := []struct {
		name           string
		blockHash      string
		mockResponse   *domain.BlockResponse
		expectedResult interface{}
		expectedError  map[string]interface{}
	}{
		{
			name:      "Success with transactions",
			blockHash: "0x123abc",
			mockResponse: &domain.BlockResponse{
				Count: 5,
			},
			expectedResult: "0x5",
			expectedError:  nil,
		},
		{
			name:           "Block not found",
			blockHash:      "0xnonexistent",
			mockResponse:   nil,
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:      "Zero transactions",
			blockHash: "0x456def",
			mockResponse: &domain.BlockResponse{
				Count: 0,
			},
			expectedResult: "0x0",
			expectedError:  nil,
		},
		{
			name:      "Large number of transactions",
			blockHash: "0x789ghi",
			mockResponse: &domain.BlockResponse{
				Count: 1000,
			},
			expectedResult: "0x3e8", // 1000 in hex
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up cache expectations
			cacheService.EXPECT().
				Get(gomock.Any(), fmt.Sprintf("eth_getBlockTransactionCountByHash_%s", tc.blockHash), gomock.Any()).
				Return(fmt.Errorf("not found"))

			mockClient.EXPECT().
				GetBlockByHashOrNumber(tc.blockHash).
				Return(tc.mockResponse)

			if tc.mockResponse != nil {
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getBlockTransactionCountByHash_%s", tc.blockHash), tc.expectedResult, service.DefaultExpiration).
					Return(nil)
			}

			result, errMap := s.GetBlockTransactionCountByHash(tc.blockHash)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, errMap)
		})
	}
}

func TestGetGasPrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create a cache service for testing
	cacheService := mocks.NewMockCacheService(ctrl)

	// Create mock client
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(
		nil,        // hClient not needed for this test
		mockClient, // pass the mock as the interface
		logger,
		nil, // tieredLimiter not needed for this test
		defaultChainId,
		cacheService,
	)

	// Set up cache expectations
	cacheService.EXPECT().
		Get(gomock.Any(), "eth_gasPrice", gomock.Any()).
		Return(fmt.Errorf("not found"))

	mockClient.EXPECT().
		GetNetworkFees("", "").
		Return(int64(100), nil) // Return 100 tinybars

	expectedResult := "0xe8d4a51000"
	cacheService.EXPECT().
		Set(gomock.Any(), "eth_gasPrice", expectedResult, service.DefaultExpiration).
		Return(nil)

	result, errMap := s.GetGasPrice()
	assert.Nil(t, errMap)

	// Expected calculation:
	// 100 tinybars * 10^10 (conversion to weibars) = 1000000000000
	// Convert to hex = 0x28fa6ae00
	assert.Equal(t, expectedResult, result)
}

func TestGetGasPrice_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

	// Set up cache expectations
	cacheService.EXPECT().
		Get(gomock.Any(), "eth_gasPrice", gomock.Any()).
		Return(fmt.Errorf("not found"))

	// Set up mirror client expectations to return error
	mockClient.EXPECT().
		GetNetworkFees("", "").
		Return(int64(0), fmt.Errorf("failed to fetch network fees"))

	result, errMap := s.GetGasPrice()
	assert.Nil(t, result)
	assert.Equal(t, map[string]interface{}{
		"code":    -32000,
		"message": "Failed to fetch gas price",
	}, errMap)
}

func TestGetChainId(t *testing.T) {
	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create mock client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheService := mocks.NewMockCacheService(ctrl)
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
				cacheService,
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
	cacheService := mocks.NewMockCacheService(ctrl)

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
			// Set up cache expectations
			cacheService.EXPECT().
				Get(gomock.Any(), fmt.Sprintf("eth_getBlockByHash_%s_%t", tc.hash, tc.showDetails), gomock.Any()).
				Return(fmt.Errorf("not found"))

			mockClient.EXPECT().
				GetBlockByHashOrNumber(tc.hash).
				Return(tc.mockResponse)

			if tc.mockResponse != nil {
				mockClient.EXPECT().
					GetContractResults(tc.mockResponse.Timestamp).
					Return(tc.mockResults)

				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getBlockByHash_%s_%t", tc.hash, tc.showDetails), gomock.Any(), service.DefaultExpiration).
					Return(nil)
			}

			s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)
			result, errMap := s.GetBlockByHash(tc.hash, tc.showDetails)

			if tc.expectNil {
				assert.Nil(t, result)
				assert.Nil(t, errMap)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, errMap)

				block, ok := result.(*domain.Block)
				assert.True(t, ok, "Result should be of type *domain.Block")
				assert.Equal(t, fmt.Sprintf("0x%x", tc.mockResponse.Number), *block.Number)
				assert.Equal(t, tc.mockResponse.Hash, *block.Hash)
				assert.Equal(t, tc.mockResponse.PreviousHash, block.ParentHash)
				assert.Equal(t, fmt.Sprintf("0x%x", tc.mockResponse.GasUsed), block.GasUsed)
				assert.Equal(t, fmt.Sprintf("0x%x", tc.mockResponse.Size), block.Size)
				assert.Equal(t, tc.mockResponse.LogsBloom, block.LogsBloom)
				if tc.showDetails {
					assert.Equal(t, len(tc.mockResults), len(block.Transactions))
				} else {
					assert.Equal(t, len(tc.mockResults), len(block.Transactions))
					for i, tx := range tc.mockResults {
						assert.Equal(t, tx.Hash, block.Transactions[i])
					}
				}
			}
		})
	}
}

func TestGetBlockByNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

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
		name         string
		numberOrTag  string
		showDetails  bool
		mockResponse *domain.BlockResponse
		mockResults  []domain.ContractResults
		expectNil    bool
		setupMocks   func()
	}{
		{
			name:         "Success with specific number",
			numberOrTag:  "0x7b",
			showDetails:  false,
			mockResponse: expectedBlock,
			mockResults:  []domain.ContractResults{{Hash: "0xtx1"}},
			expectNil:    false,
			setupMocks: func() {
				cacheKey := fmt.Sprintf("%s_%s_%t", service.GetBlockByNumber, "0x7b", false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("123").
					Return(expectedBlock)

				mockClient.EXPECT().
					GetContractResults(expectedBlock.Timestamp).
					Return([]domain.ContractResults{{Hash: "0xtx1"}})

				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
		},
		{
			name:         "Success with latest tag",
			numberOrTag:  "latest",
			showDetails:  false,
			mockResponse: expectedBlock,
			mockResults:  []domain.ContractResults{{Hash: "0xtx1"}},
			expectNil:    false,
			setupMocks: func() {
				cacheKey := fmt.Sprintf("%s_%s_%t", service.GetBlockByNumber, "latest", false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(123)}, nil)

				mockClient.EXPECT().
					GetBlockByHashOrNumber("123").
					Return(expectedBlock)

				mockClient.EXPECT().
					GetContractResults(expectedBlock.Timestamp).
					Return([]domain.ContractResults{{Hash: "0xtx1"}})

				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
		},
		{
			name:         "Success with earliest tag",
			numberOrTag:  "earliest",
			showDetails:  false,
			mockResponse: expectedBlock,
			mockResults:  []domain.ContractResults{{Hash: "0xtx1"}},
			expectNil:    false,
			setupMocks: func() {
				cacheKey := fmt.Sprintf("%s_%s_%t", service.GetBlockByNumber, "earliest", false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("0").
					Return(expectedBlock)

				mockClient.EXPECT().
					GetContractResults(expectedBlock.Timestamp).
					Return([]domain.ContractResults{{Hash: "0xtx1"}})

				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
		},
		{
			name:         "Block not found",
			numberOrTag:  "0x999",
			showDetails:  false,
			mockResponse: nil,
			expectNil:    true,
			setupMocks: func() {
				cacheKey := fmt.Sprintf("%s_%s_%t", service.GetBlockByNumber, "0x999", false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("2457").
					Return(nil)
			},
		},
		{
			name:        "Invalid hex number",
			numberOrTag: "0xinvalid",
			showDetails: false,
			expectNil:   false,
			setupMocks: func() {
				cacheKey := fmt.Sprintf("%s_%s_%t", service.GetBlockByNumber, "0xinvalid", false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))
			},
		},
		{
			name:         "Success with cached block",
			numberOrTag:  "0x7b",
			showDetails:  false,
			mockResponse: expectedBlock,
			expectNil:    false,
			setupMocks: func() {
				cacheKey := fmt.Sprintf("%s_%s_%t", service.GetBlockByNumber, "0x7b", false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, block interface{}) error {
						b := block.(*domain.Block)
						hexNum := "0x7b"
						hexHash := expectedBlock.Hash
						b.Number = &hexNum
						b.Hash = &hexHash
						b.ParentHash = expectedBlock.PreviousHash
						b.LogsBloom = expectedBlock.LogsBloom
						b.TransactionsRoot = &hexHash
						b.GasUsed = "0x3e8"     // 1000 in hex
						b.Size = "0x7d0"        // 2000 in hex
						b.GasLimit = "0xe4e1c0" // Default gas limit
						b.Nonce = "0x0000000000000000"
						b.Sha3Uncles = "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"
						b.StateRoot = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
						b.ReceiptsRoot = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
						b.Miner = "0x0000000000000000000000000000000000000000"
						b.Difficulty = "0x0"
						b.ExtraData = "0x"
						b.Timestamp = "0x61cf9980"       // Adding timestamp field
						b.Transactions = []interface{}{} // Empty transactions array
						b.Uncles = []string{}            // Empty uncles array
						return nil
					})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)
			result, errMap := s.GetBlockByNumber(tc.numberOrTag, tc.showDetails)

			if tc.name == "Invalid hex number" {
				assert.NotNil(t, errMap)
				assert.Equal(t, -32000, errMap["code"])
				assert.Equal(t, "Failed to parse hex value", errMap["message"])
				return
			}

			if tc.expectNil {
				assert.Nil(t, result)
				assert.Nil(t, errMap)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, errMap)

				block, ok := result.(*domain.Block)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, "0x7b", *block.Number)
					assert.Equal(t, expectedBlock.Hash, *block.Hash)
					assert.Equal(t, expectedBlock.PreviousHash, block.ParentHash)
					if !strings.Contains(tc.name, "cached") {
						assert.Equal(t, len(tc.mockResults), len(block.Transactions))
					}
				}
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	cacheService := mocks.NewMockCacheService(ctrl)
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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

	cacheService := mocks.NewMockCacheService(ctrl)

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
		cacheService,
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

	cacheService := mocks.NewMockCacheService(ctrl)

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
		cacheService,
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

	cacheService := mocks.NewMockCacheService(ctrl)

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
		cacheService,
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

	cacheService := mocks.NewMockCacheService(ctrl)

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
		cacheService,
	)

	result := s.GetBalance("0x123", "999999")
	assert.Equal(t, "0x0", result)
}

func TestCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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
			expectedResult: "0x0000000000000000000000000000000000000000000000000000000000000064",
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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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
			// Set up cache expectations
			var cachedTx interface{}
			cacheService.EXPECT().
				Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByHash_%s", tc.hash), &cachedTx).
				Return(errors.New("not found")).
				Times(1)

			mockClient.EXPECT().
				GetContractResult(tc.hash).
				Return(tc.mockResult).
				Times(1)

			if tc.mockResult != nil {
				result := tc.mockResult.(domain.ContractResultResponse)
				mockClient.EXPECT().GetContractById(result.To).Return(nil, nil).AnyTimes()
				mockClient.EXPECT().GetContractById(result.From).Return(nil, nil).AnyTimes()
				mockClient.EXPECT().GetAccountById(result.To).Return(nil, nil).AnyTimes()
				mockClient.EXPECT().GetAccountById(result.From).Return(nil, nil).AnyTimes()

				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionByHash_%s", tc.hash), gomock.Any(), service.DefaultExpiration).
					Return(nil).
					Times(1)
			}

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
	cacheService := mocks.NewMockCacheService(ctrl)
	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, "0x12a", cacheService)

	txHash := "0x123"
	blockHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	blockTimestamp := domain.Timestamp{
		From: "123",
		To:   "456",
	}

	testCases := []struct {
		name        string
		hash        string
		mockResult  interface{}
		mockBlock   *domain.BlockResponse
		mockFee     int64
		checkFields func(t *testing.T, result interface{})
	}{
		{
			name: "successful transaction receipt",
			hash: txHash,
			mockResult: domain.ContractResultResponse{
				BlockHash:        blockHash,
				BlockNumber:      123,
				Hash:             txHash,
				From:             "0xabc",
				To:               "0xdef",
				GasUsed:          100000,
				Status:           "0x1",
				TransactionIndex: 1,
				Type:             nil,
				Logs:             []domain.MirroNodeLogs{},
			},
			mockBlock: &domain.BlockResponse{
				Hash:      blockHash,
				Timestamp: blockTimestamp,
			},
			mockFee: 100000,
			checkFields: func(t *testing.T, result interface{}) {
				receipt, ok := result.(domain.TransactionReceipt)
				assert.True(t, ok)
				assert.Equal(t, blockHash, receipt.BlockHash)
				assert.Equal(t, "0x7b", receipt.BlockNumber)
				assert.Equal(t, "0x1", receipt.TransactionIndex)
				// 100000 * 10^10 = 1000000000000000 (0x38d7ea4c68000)
				assert.Equal(t, "0x38d7ea4c68000", receipt.EffectiveGasPrice)
			},
		},
		{
			name:       "transaction not found",
			hash:       "0xnonexistent",
			mockResult: nil,
			mockBlock:  nil,
			mockFee:    0,
			checkFields: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "with logs",
			hash: txHash,
			mockResult: domain.ContractResultResponse{
				BlockHash:        blockHash,
				BlockNumber:      123,
				Hash:             txHash,
				From:             "0xabc",
				To:               "0xdef",
				GasUsed:          100000,
				Status:           "0x1",
				TransactionIndex: 1,
				Type:             nil,
				Logs: []domain.MirroNodeLogs{
					{
						Address: "0xlog1",
						Data:    "0xdata1",
						Topics:  []string{"0xtopic1", "0xtopic2"},
					},
				},
				Bloom: "0x1234",
			},
			mockBlock: &domain.BlockResponse{
				Hash:      blockHash,
				Timestamp: blockTimestamp,
			},
			mockFee: 100000,
			checkFields: func(t *testing.T, result interface{}) {
				receipt, ok := result.(domain.TransactionReceipt)
				assert.True(t, ok)
				assert.Equal(t, "0x1234", receipt.LogsBloom)
				assert.Len(t, receipt.Logs, 1)
				assert.Equal(t, "0xlog1", receipt.Logs[0].Address)
				assert.Equal(t, "0xdata1", receipt.Logs[0].Data)
				assert.Equal(t, []string{"0xtopic1", "0xtopic2"}, receipt.Logs[0].Topics)
			},
		},
		{
			name: "with empty bloom",
			hash: txHash,
			mockResult: domain.ContractResultResponse{
				BlockHash:        blockHash,
				BlockNumber:      123,
				Hash:             txHash,
				From:             "0xabc",
				To:               "0xdef",
				GasUsed:          100000,
				Status:           "0x1",
				TransactionIndex: 1,
				Type:             nil,
				Bloom:            "0x",
			},
			mockBlock: &domain.BlockResponse{
				Hash:      blockHash,
				Timestamp: blockTimestamp,
			},
			mockFee: 100000,
			checkFields: func(t *testing.T, result interface{}) {
				receipt, ok := result.(domain.TransactionReceipt)
				assert.True(t, ok)
				assert.Equal(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", receipt.LogsBloom)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up cache expectations
			var cachedTx interface{}
			cacheService.EXPECT().
				Get(gomock.Any(), fmt.Sprintf("eth_getTransactionReceipt_%s", tc.hash), &cachedTx).
				Return(errors.New("not found")).
				Times(1)

			mockClient.EXPECT().
				GetContractResult(tc.hash).
				Return(tc.mockResult).
				Times(1)

			if tc.mockResult != nil {
				result := tc.mockResult.(domain.ContractResultResponse)
				mockClient.EXPECT().GetContractById(result.To).Return(nil, nil).AnyTimes()
				mockClient.EXPECT().GetContractById(result.From).Return(nil, nil).AnyTimes()
				mockClient.EXPECT().GetAccountById(result.To).Return(nil, nil).AnyTimes()
				mockClient.EXPECT().GetAccountById(result.From).Return(nil, nil).AnyTimes()

				mockClient.EXPECT().
					GetBlockByHashOrNumber(tc.mockBlock.Hash).
					Return(tc.mockBlock).
					Times(1)

				mockClient.EXPECT().
					GetNetworkFees(tc.mockBlock.Timestamp.From, gomock.Any()).
					Return(tc.mockFee, nil).
					Times(1)

				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionReceipt_%s", tc.hash), gomock.Any(), service.DefaultExpiration).
					Return(nil).
					Times(1)
			}

			result, errMap := s.GetTransactionReceipt(tc.hash)
			assert.Nil(t, errMap)
			if tc.checkFields != nil {
				tc.checkFields(t, result)
			}
		})
	}
}

func TestFeeHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

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
			name:              "Success_with_fixed_fee",
			blockCount:        "0x5",
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
					}, nil).
					Times(2)

				cacheService.EXPECT().
					Get(gomock.Any(), GetGasPrice, gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				mockClient.EXPECT().
					GetNetworkFees("", "").
					Return(int64(10000000000), nil).
					Times(1)

				cacheService.EXPECT().
					Set(gomock.Any(), GetGasPrice, gomock.Any(), service.DefaultExpiration).
					Return(nil).
					Times(1)
			},
			validateResult: func(t *testing.T, result interface{}) {
				feeHistory, ok := result.(*domain.FeeHistory)
				assert.True(t, ok)
				assert.Equal(t, fmt.Sprintf("0x%x", 96), feeHistory.OldestBlock)
				assert.Equal(t, 6, len(feeHistory.BaseFeePerGas))
				assert.Equal(t, 5, len(feeHistory.GasUsedRatio))
				assert.Equal(t, [][]string(nil), feeHistory.Reward)
			},
		},
		{
			name:              "Success_with_cached_gas_price",
			blockCount:        "0x3",
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
					}, nil).
					Times(2)

				cacheService.EXPECT().
					Get(gomock.Any(), GetGasPrice, gomock.Any()).
					SetArg(2, "0xf4240").
					Return(nil).
					Times(1)
			},
			validateResult: func(t *testing.T, result interface{}) {
				feeHistory, ok := result.(*domain.FeeHistory)
				assert.True(t, ok)
				assert.Equal(t, fmt.Sprintf("0x%x", 98), feeHistory.OldestBlock)
				assert.Equal(t, 4, len(feeHistory.BaseFeePerGas))
				assert.Equal(t, 3, len(feeHistory.GasUsedRatio))
				assert.Equal(t, [][]string(nil), feeHistory.Reward)
			},
		},
		{
			name:              "Invalid_block_count",
			blockCount:        "0xinvalid",
			newestBlock:       "latest",
			rewardPercentiles: []string{},
			mockLatestBlock: map[string]interface{}{
				"number": float64(100),
			},
			expectNil:   false,
			expectError: true,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": float64(100),
					}, nil).
					Times(2)
			},
		},
		{
			name:              "Invalid_newest_block",
			blockCount:        "0x5",
			newestBlock:       "0xinvalid",
			rewardPercentiles: []string{},
			mockLatestBlock: map[string]interface{}{
				"number": float64(100),
			},
			expectNil:   false,
			expectError: true,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": float64(100),
					}, nil).
					Times(1)
			},
		},
		{
			name:              "Failed_to_get_latest_block",
			blockCount:        "0x5",
			newestBlock:       "latest",
			rewardPercentiles: []string{},
			mockLatestBlock:   nil,
			expectNil:         false,
			expectError:       true,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(nil, errors.New("failed to get latest block")).
					Times(1)
			},
		},
		{
			name:              "Failed_to_get_gas_price",
			blockCount:        "0x5",
			newestBlock:       "latest",
			rewardPercentiles: []string{},
			mockLatestBlock: map[string]interface{}{
				"number": float64(100),
			},
			expectNil:   false,
			expectError: true,
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": float64(100),
					}, nil).
					Times(2)

				cacheService.EXPECT().
					Get(gomock.Any(), GetGasPrice, gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				mockClient.EXPECT().
					GetNetworkFees("", "").
					Return(int64(0), errors.New("failed to get gas price")).
					Times(1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := service.NewEthService(nil, mockClient, logger, nil, "0x12a", cacheService)

			tc.setupMocks()

			result, errMap := s.FeeHistory(tc.blockCount, tc.newestBlock, tc.rewardPercentiles)

			if tc.expectError {
				assert.NotNil(t, errMap)
				return
			}

			if tc.expectNil {
				assert.Nil(t, result)
				return
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
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

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

func TestGetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

	testCases := []struct {
		name           string
		logParams      domain.LogParams
		setupMocks     func()
		expectedResult interface{}
		expectError    bool
	}{
		{
			name: "Success with block hash",
			logParams: domain.LogParams{
				BlockHash: "0x123abc",
				Address:   []string{"0x742d35Cc6634C0532925a3b844Bc454e4438f44e"},
				Topics:    []string{"0xtopic1", "0xtopic2"},
			},
			setupMocks: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0x123abc").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							From: "2023-01-01T00:00:00.000Z",
							To:   "2023-01-01T00:00:01.000Z",
						},
					})

				expectedParams := map[string]interface{}{
					"timestamp": "gte:2023-01-01T00:00:00.000Z&timestamp=lte:2023-01-01T00:00:01.000Z",
					"topic0":    "0xtopic1",
					"topic1":    "0xtopic2",
				}

				mockClient.EXPECT().
					GetContractResultsLogsByAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e", expectedParams).
					Return([]domain.ContractResults{
						{
							Address:          "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
							BlockHash:        "0x123abc",
							BlockNumber:      100,
							Result:           "0xdata",
							Hash:             "0xtxhash",
							TransactionIndex: 1,
						},
					}, nil)
			},
			expectedResult: []domain.Log{
				{
					Address:          "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
					BlockHash:        "0x123abc",
					BlockNumber:      "0x64", // 100 in hex
					Data:             "0xdata",
					TransactionHash:  "0xtxhash",
					TransactionIndex: "1",
				},
			},
			expectError: false,
		},
		{
			name: "Success with block range within limits",
			logParams: domain.LogParams{
				FromBlock: "0x1",
				ToBlock:   "0x2",
				Address:   []string{},
			},
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)

				mockClient.EXPECT().
					GetBlockByHashOrNumber("1").
					Return(&domain.BlockResponse{
						Number: 1,
						Timestamp: domain.Timestamp{
							From: "2023-01-01T00:00:00.000Z",
						},
					})

				mockClient.EXPECT().
					GetBlockByHashOrNumber("2").
					Return(&domain.BlockResponse{
						Number: 2,
						Timestamp: domain.Timestamp{
							To: "2023-01-01T00:00:02.000Z",
						},
					})

				expectedParams := map[string]interface{}{
					"timestamp": "gte:2023-01-01T00:00:00.000Z&timestamp=lte:2023-01-01T00:00:02.000Z",
				}

				mockClient.EXPECT().
					GetContractResultsLogsWithRetry(expectedParams).
					Return([]domain.ContractResults{
						{
							BlockHash:        "0xblockhash",
							BlockNumber:      1,
							Result:           "0xdata",
							Hash:             "0xtxhash",
							TransactionIndex: 0,
						},
					}, nil)
			},
			expectedResult: []domain.Log{
				{
					BlockHash:        "0xblockhash",
					BlockNumber:      "0x1",
					Data:             "0xdata",
					TransactionHash:  "0xtxhash",
					TransactionIndex: "0",
				},
			},
			expectError: false,
		},
		{
			name: "Block range too large",
			logParams: domain.LogParams{
				FromBlock: "0x1",
				ToBlock:   "0x64", // 100 in hex
			},
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)

				mockClient.EXPECT().
					GetBlockByHashOrNumber("1").
					Return(&domain.BlockResponse{Number: 1})

				mockClient.EXPECT().
					GetBlockByHashOrNumber("100").
					Return(&domain.BlockResponse{Number: 100})
			},
			expectedResult: []domain.Log{},
			expectError:    false,
		},
		{
			name: "Invalid block hash",
			logParams: domain.LogParams{
				BlockHash: "0xinvalid",
			},
			setupMocks: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0xinvalid").
					Return(nil)
			},
			expectedResult: []domain.Log{},
			expectError:    false,
		},
		{
			name: "Latest block fetch failure",
			logParams: domain.LogParams{
				FromBlock: "0x1",
				ToBlock:   "latest",
			},
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(nil, fmt.Errorf("failed to fetch latest block"))
			},
			expectedResult: []domain.Log{},
			expectError:    false,
		},
		{
			name: "From block greater than to block",
			logParams: domain.LogParams{
				FromBlock: "0x10",
				ToBlock:   "0x5",
			},
			setupMocks: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)

				// Add expectations for block number conversions
				mockClient.EXPECT().
					GetBlockByHashOrNumber("16"). // 0x10 in decimal
					Return(&domain.BlockResponse{Number: 16})

				mockClient.EXPECT().
					GetBlockByHashOrNumber("5").
					Return(&domain.BlockResponse{Number: 5})
			},
			expectedResult: []domain.Log{},
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			result, errMap := s.GetLogs(tc.logParams)

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

func TestGetBlockTransactionCountByNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

	testCases := []struct {
		name            string
		blockParam      string
		mockLatestBlock map[string]interface{}
		mockResponse    *domain.BlockResponse
		expectedResult  interface{}
		expectedError   map[string]interface{}
		setupMock       func()
	}{
		{
			name:       "Success with specific block number",
			blockParam: "0x7b", // 123 in hex
			mockResponse: &domain.BlockResponse{
				Count: 5,
			},
			expectedResult: "0x5",
			expectedError:  nil,
			setupMock: func() {
				// Mock cache get attempt
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_getBlockTransactionCountByNumber_123", gomock.Any()).
					Return(fmt.Errorf("not found"))

				// Mock getting block data
				mockClient.EXPECT().
					GetBlockByHashOrNumber("123").
					Return(&domain.BlockResponse{Count: 5})

				// Mock cache set with the result
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_getBlockTransactionCountByNumber_123", "0x5", gomock.Any()).
					Return(nil)
			},
		},
		{
			name:       "Success with latest tag",
			blockParam: "latest",
			mockLatestBlock: map[string]interface{}{
				"number": float64(100),
			},
			mockResponse: &domain.BlockResponse{
				Count: 10,
			},
			expectedResult: "0xa",
			expectedError:  nil,
			setupMock: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)

				// Mock cache get attempt
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_getBlockTransactionCountByNumber_100", gomock.Any()).
					Return(fmt.Errorf("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("100").
					Return(&domain.BlockResponse{Count: 10})

				// Mock cache set with the result
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_getBlockTransactionCountByNumber_100", "0xa", gomock.Any()).
					Return(nil)
			},
		},
		{
			name:       "Success with earliest tag",
			blockParam: "earliest",
			mockResponse: &domain.BlockResponse{
				Count: 1,
			},
			expectedResult: "0x1",
			expectedError:  nil,
			setupMock: func() {
				// Mock cache get attempt
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_getBlockTransactionCountByNumber_0", gomock.Any()).
					Return(fmt.Errorf("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("0").
					Return(&domain.BlockResponse{Count: 1})

				// Mock cache set with the result
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_getBlockTransactionCountByNumber_0", "0x1", gomock.Any()).
					Return(nil)
			},
		},
		{
			name:           "Block not found",
			blockParam:     "0x999",
			mockResponse:   nil,
			expectedResult: nil,
			expectedError:  nil,
			setupMock: func() {
				// Mock cache get attempt
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_getBlockTransactionCountByNumber_2457", gomock.Any()).
					Return(fmt.Errorf("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("2457"). // 0x999 in decimal
					Return(nil)
			},
		},
		{
			name:           "Invalid block number format",
			blockParam:     "0xinvalid",
			expectedResult: nil,
			expectedError: map[string]interface{}{
				"code":    -32000,
				"message": "Failed to parse hex value",
			},
			setupMock: func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, errMap := s.GetBlockTransactionCountByNumber(tc.blockParam)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, errMap)
		})
	}
}

func TestGetTransactionByBlockHashAndIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

	testBlockHash := "0x" + strings.Repeat("1", 64)
	baseContractResult := domain.ContractResults{
		BlockNumber:        123,
		BlockHash:          testBlockHash,
		Hash:               "0x" + strings.Repeat("a", 64),
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
		Type:               0,
	}

	// Mock responses for contract and account lookups
	mockContractResponse := &domain.ContractResponse{
		EvmAddress: "0x" + strings.Repeat("3", 40),
	}
	mockAccountResponse := &domain.AccountResponse{
		EvmAddress: "0x" + strings.Repeat("2", 40),
	}

	testCases := []struct {
		name           string
		blockHash      string
		index          string
		mockResult     *domain.ContractResults
		expectedResult interface{}
		expectedError  map[string]interface{}
		setupMocks     func()
		checkFields    func(t *testing.T, result interface{})
	}{
		{
			name:       "Success - Legacy transaction (type 0)",
			blockHash:  testBlockHash,
			index:      "0x1",
			mockResult: &baseContractResult,
			setupMocks: func() {
				// Mock cache miss
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockHashAndIndex_%s_%s", testBlockHash, "0x1"), gomock.Any()).
					Return(fmt.Errorf("not found"))

				// Mock contract result lookup
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(&baseContractResult, nil)

				// Mock address resolution for 'to' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.To).
					Return(mockContractResponse, nil)

				// Mock address resolution for 'from' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.From).
					Return(nil, nil)
				mockClient.EXPECT().
					GetAccountById(baseContractResult.From).
					Return(mockAccountResponse, nil)

				// Mock cache set
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockHashAndIndex_%s_%s", testBlockHash, "0x1"), gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction)
				assert.True(t, ok)
				assert.Equal(t, "0x0", tx.Type)
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x7b", *tx.BlockNumber) // 123 in hex
				assert.Equal(t, defaultChainId, *tx.ChainId)
				assert.Equal(t, "0x1", *tx.TransactionIndex)
				assert.Equal(t, mockContractResponse.EvmAddress, *tx.To)
				assert.Equal(t, mockAccountResponse.EvmAddress, tx.From)
			},
		},
		{
			name:           "Invalid transaction index",
			blockHash:      testBlockHash,
			index:          "0xinvalid",
			mockResult:     nil,
			expectedResult: nil,
			expectedError: map[string]interface{}{
				"code":    -32000,
				"message": "Failed to parse hex value",
			},
			setupMocks: func() {
				// Mock cache miss - we expect cache check even for invalid input
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockHashAndIndex_%s_%s", testBlockHash, "0xinvalid"), gomock.Any()).
					Return(fmt.Errorf("not found"))
			},
		},
		{
			name:           "Transaction not found",
			blockHash:      testBlockHash,
			index:          "0x5",
			mockResult:     nil,
			expectedResult: nil,
			expectedError:  nil,
			setupMocks: func() {
				// Mock cache miss
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockHashAndIndex_%s_%s", testBlockHash, "0x5"), gomock.Any()).
					Return(fmt.Errorf("not found"))

				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(nil, nil)
			},
		},
		{
			name:      "Different transaction type",
			blockHash: testBlockHash,
			index:     "0x1",
			mockResult: func() *domain.ContractResults {
				result := baseContractResult
				result.Type = 2 // EIP-1559 transaction
				return &result
			}(),
			setupMocks: func() {
				// Mock cache miss
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockHashAndIndex_%s_%s", testBlockHash, "0x1"), gomock.Any()).
					Return(fmt.Errorf("not found"))

				// Mock contract result lookup
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(func() *domain.ContractResults {
						result := baseContractResult
						result.Type = 2
						return &result
					}(), nil)

				// Mock address resolution for 'to' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.To).
					Return(mockContractResponse, nil)

				// Mock address resolution for 'from' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.From).
					Return(nil, nil)
				mockClient.EXPECT().
					GetAccountById(baseContractResult.From).
					Return(mockAccountResponse, nil)

				// Mock cache set
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockHashAndIndex_%s_%s", testBlockHash, "0x1"), gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction1559)
				assert.True(t, ok)
				assert.Equal(t, "0x2", tx.Type)
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x7b", *tx.BlockNumber) // 123 in hex
				assert.Equal(t, defaultChainId, *tx.ChainId)
				assert.Equal(t, mockContractResponse.EvmAddress, *tx.To)
				assert.Equal(t, mockAccountResponse.EvmAddress, tx.From)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			result, errMap := s.GetTransactionByBlockHashAndIndex(tc.blockHash, tc.index)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, errMap)
			} else {
				assert.Nil(t, errMap)
				if tc.checkFields != nil {
					tc.checkFields(t, result)
				} else {
					assert.Equal(t, tc.expectedResult, result)
				}
			}
		})
	}
}

func TestGetTransactionByBlockNumberAndIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)
	s := service.NewEthService(nil, mockClient, logger, nil, defaultChainId, cacheService)

	baseContractResult := domain.ContractResults{
		BlockNumber:        123,
		BlockHash:          "0x" + strings.Repeat("1", 64),
		Hash:               "0x" + strings.Repeat("a", 64),
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
		Type:               0,
	}

	// Mock responses for contract and account lookups
	mockContractResponse := &domain.ContractResponse{
		EvmAddress: "0x" + strings.Repeat("3", 40),
	}
	mockAccountResponse := &domain.AccountResponse{
		EvmAddress: "0x" + strings.Repeat("2", 40),
	}

	testCases := []struct {
		name           string
		blockNumber    string
		index          string
		mockResult     *domain.ContractResults
		expectedResult interface{}
		expectedError  map[string]interface{}
		setupMocks     func()
		checkFields    func(t *testing.T, result interface{})
	}{
		{
			name:        "Success with latest block",
			blockNumber: "latest",
			index:       "0x1",
			mockResult:  &baseContractResult,
			setupMocks: func() {
				// Mock cache miss
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "latest", "0x1"), gomock.Any()).
					Return(fmt.Errorf("not found"))

				// Mock getting latest block
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(123)}, nil)

				// Mock contract result lookup
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(&baseContractResult, nil)

				// Mock address resolution for 'to' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.To).
					Return(mockContractResponse, nil)

				// Mock address resolution for 'from' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.From).
					Return(nil, nil)
				mockClient.EXPECT().
					GetAccountById(baseContractResult.From).
					Return(mockAccountResponse, nil)

				// Mock cache set
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "latest", "0x1"), gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction)
				assert.True(t, ok)
				assert.Equal(t, "0x0", tx.Type)
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x7b", *tx.BlockNumber) // 123 in hex
				assert.Equal(t, defaultChainId, *tx.ChainId)
				assert.Equal(t, "0x1", *tx.TransactionIndex)
				assert.Equal(t, mockContractResponse.EvmAddress, *tx.To)
				assert.Equal(t, mockAccountResponse.EvmAddress, tx.From)
			},
		},
		{
			name:        "Success with earliest block",
			blockNumber: "earliest",
			index:       "0x1",
			mockResult:  &baseContractResult,
			setupMocks: func() {
				// Mock cache miss
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "earliest", "0x1"), gomock.Any()).
					Return(fmt.Errorf("not found"))

				// Mock contract result lookup
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(&baseContractResult, nil)

				// Mock address resolution for 'to' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.To).
					Return(mockContractResponse, nil)

				// Mock address resolution for 'from' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.From).
					Return(nil, nil)
				mockClient.EXPECT().
					GetAccountById(baseContractResult.From).
					Return(mockAccountResponse, nil)

				// Mock cache set
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "earliest", "0x1"), gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction)
				assert.True(t, ok)
				assert.Equal(t, "0x0", tx.Type)
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x7b", *tx.BlockNumber)
				assert.Equal(t, defaultChainId, *tx.ChainId)
				assert.Equal(t, mockContractResponse.EvmAddress, *tx.To)
				assert.Equal(t, mockAccountResponse.EvmAddress, tx.From)
			},
		},
		{
			name:           "Invalid transaction index",
			blockNumber:    "0x7b", // 123 in hex
			index:          "0xinvalid",
			mockResult:     nil,
			expectedResult: nil,
			expectedError: map[string]interface{}{
				"code":    -32000,
				"message": "Failed to parse hex value",
			},
			setupMocks: func() {
				// Mock cache miss - we expect cache check even for invalid input
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "0x7b", "0xinvalid"), gomock.Any()).
					Return(fmt.Errorf("not found"))
			},
		},
		{
			name:           "Transaction not found",
			blockNumber:    "0x7b", // 123 in hex
			index:          "0x5",
			mockResult:     nil,
			expectedResult: nil,
			expectedError:  nil,
			setupMocks: func() {
				// Mock cache miss
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "0x7b", "0x5"), gomock.Any()).
					Return(fmt.Errorf("not found"))

				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(nil, nil)
			},
		},
		{
			name:        "Different transaction type",
			blockNumber: "0x7b", // 123 in hex
			index:       "0x1",
			mockResult: func() *domain.ContractResults {
				result := baseContractResult
				result.Type = 2 // EIP-1559 transaction
				return &result
			}(),
			setupMocks: func() {
				// Mock cache miss
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "0x7b", "0x1"), gomock.Any()).
					Return(fmt.Errorf("not found"))

				// Mock contract result lookup
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(func() *domain.ContractResults {
						result := baseContractResult
						result.Type = 2
						return &result
					}(), nil)

				// Mock address resolution for 'to' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.To).
					Return(mockContractResponse, nil)

				// Mock address resolution for 'from' address
				mockClient.EXPECT().
					GetContractById(baseContractResult.From).
					Return(nil, nil)
				mockClient.EXPECT().
					GetAccountById(baseContractResult.From).
					Return(mockAccountResponse, nil)

				// Mock cache set
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionByBlockNumberAndIndex_%s_%s", "0x7b", "0x1"), gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction1559)
				assert.True(t, ok)
				assert.Equal(t, "0x2", tx.Type)
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x7b", *tx.BlockNumber)
				assert.Equal(t, defaultChainId, *tx.ChainId)
				assert.Equal(t, mockContractResponse.EvmAddress, *tx.To)
				assert.Equal(t, mockAccountResponse.EvmAddress, tx.From)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			result, errMap := s.GetTransactionByBlockNumberAndIndex(tc.blockNumber, tc.index)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, errMap)
			} else {
				assert.Nil(t, errMap)
				if tc.checkFields != nil {
					tc.checkFields(t, result)
				} else {
					assert.Equal(t, tc.expectedResult, result)
				}
			}
		})
	}
}

func TestGetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create a cache service for testing
	cacheService := mocks.NewMockCacheService(ctrl)

	// Create mock client from the interface
	mockClient := mocks.NewMockMirrorClient(ctrl)
	mockHederaClient := mocks.NewMockHederaNodeClient(ctrl)

	s := service.NewEthService(
		mockHederaClient,
		mockClient,
		logger,
		nil,
		defaultChainId,
		cacheService,
	)

	t.Run("iHTS precompile address", func(t *testing.T) {
		address := "0x0000000000000000000000000000000000000167"
		blockNumber := "latest"

		result, errMap := s.GetCode(address, blockNumber)

		assert.Equal(t, "0xfe", result)
		assert.Nil(t, errMap)
	})

	t.Run("Contract with runtime bytecode", func(t *testing.T) {
		address := "0x123"
		blockNumber := "latest"
		runtimeBytecode := "0x606060"

		cacheKey := fmt.Sprintf("%s_%s_%s", GetCode, address, blockNumber)
		cacheService.EXPECT().
			Get(gomock.Any(), cacheKey, gomock.Any()).
			Return(errors.New("not found"))

		mockClient.EXPECT().
			GetContractById(address).
			Return(&domain.ContractResponse{
				RuntimeBytecode: &runtimeBytecode,
			}, nil)

		cacheService.EXPECT().
			Set(gomock.Any(), cacheKey, runtimeBytecode, DefaultExpiration).
			Return(nil)

		result, errMap := s.GetCode(address, blockNumber)

		assert.Equal(t, runtimeBytecode, result)
		assert.Nil(t, errMap)
	})

	t.Run("Token redirect case", func(t *testing.T) {
		address := "0x000000000000123"
		blockNumber := "latest"

		cacheKey := fmt.Sprintf("%s_%s_%s", GetCode, address, blockNumber)
		cacheService.EXPECT().
			Get(gomock.Any(), cacheKey, gomock.Any()).
			Return(errors.New("not found"))

		mockClient.EXPECT().
			GetContractById(address).
			Return(nil, nil)

		mockClient.EXPECT().
			GetAccountById(address).
			Return(nil, nil)

		mockClient.EXPECT().
			GetTokenById("0.0.291").
			Return(&domain.TokenResponse{
				TokenId: "0.0.291",
			}, nil)

		result, errMap := s.GetCode(address, blockNumber)

		expectedBytecode := "0x" + "6080604052348015600f57600080fd5b506000610167905077618dc65e" + address[2:] + "600052366000602037600080366018016008845af43d806000803e8160008114605857816000f35b816000fdfea2646970667358221220d8378feed472ba49a0005514ef7087017f707b45fb9bf56bb81bb93ff19a238b64736f6c634300080b0033"
		assert.Equal(t, expectedBytecode, result)
		assert.Nil(t, errMap)
	})

	t.Run("Fallback to Hedera client", func(t *testing.T) {
		address := "0x456"
		blockNumber := "latest"
		bytecode := []byte{1, 2, 3}

		cacheKey := fmt.Sprintf("%s_%s_%s", GetCode, address, blockNumber)
		cacheService.EXPECT().
			Get(gomock.Any(), cacheKey, gomock.Any()).
			Return(errors.New("not found"))

		mockClient.EXPECT().
			GetContractById(address).
			Return(nil, nil)

		mockClient.EXPECT().
			GetAccountById(address).
			Return(nil, nil)

		mockHederaClient.EXPECT().
			GetContractByteCode(gomock.Any(), gomock.Any(), address).
			Return(bytecode, nil)

		expectedResponse := fmt.Sprintf("0x%x", bytecode)
		cacheService.EXPECT().
			Set(gomock.Any(), cacheKey, expectedResponse, DefaultExpiration).
			Return(nil)

		result, errMap := s.GetCode(address, blockNumber)

		assert.Equal(t, expectedResponse, result)
		assert.Nil(t, errMap)
	})

	t.Run("Cache hit", func(t *testing.T) {
		address := "0x789"
		blockNumber := "latest"
		cachedBytecode := "0xabcdef"

		cacheKey := fmt.Sprintf("%s_%s_%s", GetCode, address, blockNumber)
		cacheService.EXPECT().
			Get(gomock.Any(), cacheKey, gomock.Any()).
			SetArg(2, cachedBytecode).
			Return(nil)

		result, errMap := s.GetCode(address, blockNumber)

		assert.Equal(t, cachedBytecode, result)
		assert.Nil(t, errMap)
	})

	t.Run("Hedera client error", func(t *testing.T) {
		address := "0x999"
		blockNumber := "latest"

		cacheKey := fmt.Sprintf("%s_%s_%s", GetCode, address, blockNumber)
		cacheService.EXPECT().
			Get(gomock.Any(), cacheKey, gomock.Any()).
			Return(errors.New("not found"))

		mockClient.EXPECT().
			GetContractById(address).
			Return(nil, nil)

		mockClient.EXPECT().
			GetAccountById(address).
			Return(nil, nil)

		mockHederaClient.EXPECT().
			GetContractByteCode(gomock.Any(), gomock.Any(), address).
			Return(nil, errors.New("hedera error"))

		result, errMap := s.GetCode(address, blockNumber)

		assert.Equal(t, "0x", result)
		assert.Nil(t, errMap)
	})
}

func TestSendRawTransactionEndpoint(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMirrorClient := mocks.NewMockMirrorClient(ctrl)
	mockHederaClient := mocks.NewMockHederaNodeClient(ctrl)
	mockCacheService := mocks.NewMockCacheService(ctrl)

	logger := zap.NewNop()
	ethService := service.NewEthService(mockHederaClient, mockMirrorClient, logger, nil, "0x128", mockCacheService)

	// Test case 1: Successful transaction
	t.Run("Successful transaction", func(t *testing.T) {
		// Mock cache service for gas price
		mockCacheService.EXPECT().
			Get(gomock.Any(), "eth_gasPrice", gomock.Any()).
			SetArg(2, "0x4f29944800").
			Return(nil)

		// Mock GetAccount for contract address
		mockMirrorClient.EXPECT().
			GetAccount(gomock.Any(), gomock.Any()).
			Return(nil)

		// Mock GetAccountById for sender address
		mockMirrorClient.EXPECT().
			GetAccountById(gomock.Any()).
			Return(&domain.AccountResponse{
				EvmAddress: "0x96216849c49358B10257cb55b28eA603c874b05E",
				Balance: struct {
					Balance   int64         `json:"balance"`
					Timestamp string        `json:"timestamp"`
					Tokens    []interface{} `json:"tokens"`
				}{
					Balance:   1000000000,
					Timestamp: "2021-01-01T00:00:00Z",
					Tokens:    []interface{}{},
				},
			}, nil)

		rawTxHex := "0xf8cc1e854f29944800832dc6c0940a56fd9e0c4f67df549e7f375a9451c0086482ec80b864a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000b757064617465645f6d7367000000000000000000000000000000000000000000820274a0cd6095ae91ea5d609b32923a9f73572e2d031fde0b7e38de44d3eda187474140a03028ecf5eb61070cba8e927ad5e11eac116da441307f2d54dae8be90f4476c59"

		expectedHash := "0x123456789abcdef"

		// Mock successful transaction
		mockHederaClient.EXPECT().
			SendRawTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&hedera.TransactionResponse{
				TransactionID: "0.0.1234@1234567890.123456789",
			}, nil)

		mockMirrorClient.EXPECT().
			RepeatGetContractResult(gomock.Any(), gomock.Any()).
			Return(&domain.ContractResultResponse{
				Hash: expectedHash,
			})

		result, errMap := ethService.SendRawTransaction(rawTxHex)

		assert.Nil(t, errMap)
		resultStr, ok := result.(*string)
		assert.True(t, ok)
		assert.Equal(t, expectedHash, *resultStr)
	})

	// Test case 2: Invalid transaction data
	t.Run("Invalid transaction data", func(t *testing.T) {
		result, errMap := ethService.SendRawTransaction("")

		assert.NotNil(t, errMap)
		assert.Nil(t, result)
		assert.Equal(t, -32000, errMap["code"])
		assert.Contains(t, errMap["message"], "transaction data is empty")
	})
}
