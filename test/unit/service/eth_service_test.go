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

const (
	defaultChainId = "0x12a"
	// Token redirect bytecode constants
	redirectBytecodePrefix  = "6080604052348015600f57600080fd5b506000610167905077618dc65e"
	redirectBytecodePostfix = "600052366000602037600080366018016008845af43d806000803e8160008114605857816000f35b816000fdfea2646970667358221220d8378feed472ba49a0005514ef7087017f707b45fb9bf56bb81bb93ff19a238b64736f6c634300080b0033"
)

const GetGasPrice = "eth_gasPrice"
const GetCode = "eth_getCode"
const GetBlockNumber = "eth_blockNumber"
const DefaultExpiration = time.Hour     // 1 hour expiration
const ShortExpiration = 1 * time.Second // 10 minutes expiration

// Helper functions for creating pointers
func ptr[T any](v T) *T {
	return &v
}

func TestGetBlockNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create a cache service for testing
	cacheService := mocks.NewMockCacheService(ctrl)
	commonService := mocks.NewMockCommonService(ctrl)
	mockClient := mocks.NewMockMirrorClient(ctrl)

	// Set up expectations
	cacheService.EXPECT().
		Get(gomock.Any(), GetBlockNumber, gomock.Any()).
		Return(errors.New("not found")).
		Times(1)

	commonService.EXPECT().
		GetBlockNumber().
		Return("0x2a", nil).
		Times(1)

	cacheService.EXPECT().
		Set(gomock.Any(), GetBlockNumber, "0x2a", ShortExpiration).
		Return(nil).
		Times(1)

	s := service.NewEthService(
		nil,
		mockClient,
		commonService,
		logger,
		nil,
		defaultChainId,
		cacheService,
	)

	result, errMap := s.GetBlockNumber()
	assert.Nil(t, errMap)
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
		nil,
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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

	// Test all uncle-related methods
	t.Run("GetUncleCountByBlockNumber", func(t *testing.T) {
		result, errMap := s.GetUncleCountByBlockNumber("0x1")
		assert.Nil(t, errMap)
		assert.Equal(t, "0x0", result)
	})

	t.Run("GetUncleByBlockNumberAndIndex", func(t *testing.T) {
		result, errMap := s.GetUncleByBlockNumberAndIndex("0x1", "0x0")
		assert.Nil(t, errMap)
		assert.Nil(t, result)
	})

	t.Run("GetUncleCountByBlockHash", func(t *testing.T) {
		result, errMap := s.GetUncleCountByBlockHash("0x1234567890123456789012345678901234567890123456789012345678901234")
		assert.Nil(t, errMap)
		assert.Equal(t, "0x0", result)
	})

	t.Run("GetUncleByBlockHashAndIndex", func(t *testing.T) {
		result, errMap := s.GetUncleByBlockHashAndIndex("0x1234567890123456789012345678901234567890123456789012345678901234", "0x0")
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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

	testCases := []struct {
		name           string
		blockHash      string
		mockResponse   *domain.BlockResponse
		expectedResult interface{}
		expectedError  *domain.RPCError
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

			result, errRpc := s.GetBlockTransactionCountByHash(tc.blockHash)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, errRpc)
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
		nil,
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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

	// Set up cache expectations
	cacheService.EXPECT().
		Get(gomock.Any(), "eth_gasPrice", gomock.Any()).
		Return(fmt.Errorf("not found"))

	// Set up mirror client expectations to return error
	mockClient.EXPECT().
		GetNetworkFees("", "").
		Return(int64(0), fmt.Errorf("failed to fetch network fees"))

	result, errRpc := s.GetGasPrice()
	assert.Nil(t, result)
	assert.Equal(t, domain.NewRPCError(domain.ServerError, "Failed to fetch gas price"), errRpc)
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
				nil,
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
	mockCacheService := mocks.NewMockCacheService(ctrl)

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
				{
					Hash:   "0xtx1",
					Result: "SUCCESS",
					From:   "0x" + strings.Repeat("2", 40),
					To:     "0x" + strings.Repeat("3", 40),
				},
				{
					Hash:   "0xtx2",
					Result: "SUCCESS",
					From:   "0x" + strings.Repeat("4", 40),
					To:     "0x" + strings.Repeat("5", 40),
				},
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
			// Set up cache expectations for block
			cacheKey := fmt.Sprintf("eth_getBlockByHash_%s_%t", tc.hash, tc.showDetails)
			mockCacheService.EXPECT().
				Get(gomock.Any(), cacheKey, gomock.Any()).
				Return(fmt.Errorf("not found"))

			mockClient.EXPECT().
				GetBlockByHashOrNumber(tc.hash).
				Return(tc.mockResponse)

			if tc.mockResponse != nil {
				mockClient.EXPECT().
					GetContractResults(tc.mockResponse.Timestamp).
					Return(tc.mockResults)

				// For each transaction in mockResults, set up cache expectations for resolving addresses
				for _, tx := range tc.mockResults {
					fromCacheKey := fmt.Sprintf("evm_address_%s", tx.From)
					toCacheKey := fmt.Sprintf("evm_address_%s", tx.To)

					// Mock cache Get for 'from' address
					mockCacheService.EXPECT().
						Get(gomock.Any(), fromCacheKey, gomock.Any()).
						DoAndReturn(func(_ interface{}, _ string, result *string) error {
							*result = tx.From
							return nil
						}).AnyTimes()

					// Mock cache Get for 'to' address
					mockCacheService.EXPECT().
						Get(gomock.Any(), toCacheKey, gomock.Any()).
						DoAndReturn(func(_ interface{}, _ string, result *string) error {
							*result = tx.To
							return nil
						}).AnyTimes()
				}

				mockCacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			}

			s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, mockCacheService)
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
	commonService := mocks.NewMockCommonService(ctrl)

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
			mockResults: []domain.ContractResults{{
				Hash:   "0xtx1",
				Result: "SUCCESS",
				From:   "0x" + strings.Repeat("2", 40),
				To:     "0x" + strings.Repeat("3", 40),
			}},
			expectNil: false,
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for hex block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x7b").
					Return(int64(123), nil)

				// Mock cache miss for block
				cacheKey := fmt.Sprintf("eth_getBlockByNumber_%d_%t", 123, false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock getting block data
				mockClient.EXPECT().
					GetBlockByHashOrNumber("123").
					Return(expectedBlock)

				// Mock getting contract results
				mockClient.EXPECT().
					GetContractResults(expectedBlock.Timestamp).
					Return([]domain.ContractResults{{
						Hash:   "0xtx1",
						Result: "SUCCESS",
						From:   "0x" + strings.Repeat("2", 40),
						To:     "0x" + strings.Repeat("3", 40),
					}})

				// Mock address resolution for 'from' address
				fromAddr := "0x" + strings.Repeat("2", 40)
				fromCacheKey := fmt.Sprintf("evm_address_%s", fromAddr)
				cacheService.EXPECT().
					Get(gomock.Any(), fromCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById(fromAddr).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById(fromAddr).
					Return(&domain.AccountResponse{
						EvmAddress: fromAddr,
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), fromCacheKey, fromAddr, service.DefaultExpiration).
					Return(nil)

				// Mock address resolution for 'to' address
				toAddr := "0x" + strings.Repeat("3", 40)
				toCacheKey := fmt.Sprintf("evm_address_%s", toAddr)
				cacheService.EXPECT().
					Get(gomock.Any(), toCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById(toAddr).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById(toAddr).
					Return(&domain.AccountResponse{
						EvmAddress: toAddr,
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), toCacheKey, toAddr, service.DefaultExpiration).
					Return(nil)

				// Mock cache set for block
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
			mockResults: []domain.ContractResults{{
				Hash:   "0xtx1",
				Result: "SUCCESS",
				From:   "0x" + strings.Repeat("2", 40),
				To:     "0x" + strings.Repeat("3", 40),
			}},
			expectNil: false,
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for latest block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil)

				// Mock cache miss for block
				cacheKey := fmt.Sprintf("eth_getBlockByNumber_%d_%t", 100, false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock getting block data
				mockClient.EXPECT().
					GetBlockByHashOrNumber("100").
					Return(expectedBlock)

				// Mock getting contract results
				mockClient.EXPECT().
					GetContractResults(expectedBlock.Timestamp).
					Return([]domain.ContractResults{{
						Hash:   "0xtx1",
						Result: "SUCCESS",
						From:   "0x" + strings.Repeat("2", 40),
						To:     "0x" + strings.Repeat("3", 40),
					}})

				// Mock address resolution for 'from' address
				fromAddr := "0x" + strings.Repeat("2", 40)
				fromCacheKey := fmt.Sprintf("evm_address_%s", fromAddr)
				cacheService.EXPECT().
					Get(gomock.Any(), fromCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById(fromAddr).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById(fromAddr).
					Return(&domain.AccountResponse{
						EvmAddress: fromAddr,
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), fromCacheKey, fromAddr, service.DefaultExpiration).
					Return(nil)

				// Mock address resolution for 'to' address
				toAddr := "0x" + strings.Repeat("3", 40)
				toCacheKey := fmt.Sprintf("evm_address_%s", toAddr)
				cacheService.EXPECT().
					Get(gomock.Any(), toCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById(toAddr).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById(toAddr).
					Return(&domain.AccountResponse{
						EvmAddress: toAddr,
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), toCacheKey, toAddr, service.DefaultExpiration).
					Return(nil)

				// Mock cache set for block
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
			mockResults: []domain.ContractResults{{
				Hash:   "0xtx1",
				Result: "SUCCESS",
				From:   "0x" + strings.Repeat("2", 40),
				To:     "0x" + strings.Repeat("3", 40),
			}},
			expectNil: false,
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for earliest block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("earliest").
					Return(int64(0), nil)

				// Mock cache miss for block
				cacheKey := fmt.Sprintf("eth_getBlockByNumber_%d_%t", 0, false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock getting block data
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0").
					Return(expectedBlock)

				// Mock getting contract results
				mockClient.EXPECT().
					GetContractResults(expectedBlock.Timestamp).
					Return([]domain.ContractResults{{
						Hash:   "0xtx1",
						Result: "SUCCESS",
						From:   "0x" + strings.Repeat("2", 40),
						To:     "0x" + strings.Repeat("3", 40),
					}})

				// Mock address resolution for 'from' address
				fromAddr := "0x" + strings.Repeat("2", 40)
				fromCacheKey := fmt.Sprintf("evm_address_%s", fromAddr)
				cacheService.EXPECT().
					Get(gomock.Any(), fromCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById(fromAddr).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById(fromAddr).
					Return(&domain.AccountResponse{
						EvmAddress: fromAddr,
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), fromCacheKey, fromAddr, service.DefaultExpiration).
					Return(nil)

				// Mock address resolution for 'to' address
				toAddr := "0x" + strings.Repeat("3", 40)
				toCacheKey := fmt.Sprintf("evm_address_%s", toAddr)
				cacheService.EXPECT().
					Get(gomock.Any(), toCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById(toAddr).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById(toAddr).
					Return(&domain.AccountResponse{
						EvmAddress: toAddr,
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), toCacheKey, toAddr, service.DefaultExpiration).
					Return(nil)

				// Mock cache set for block
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
				// Mock GetBlockNumberByNumberOrTag for non-existent block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x999").
					Return(int64(2457), nil)

				// Mock cache miss for block
				cacheKey := fmt.Sprintf("eth_getBlockByNumber_%d_%t", 2457, false)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock getting block data returns nil for non-existent block
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
				// Mock GetBlockNumberByNumberOrTag to return error for invalid hex
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0xinvalid").
					Return(int64(0), domain.NewRPCError(domain.ServerError, "Invalid block number"))
			},
		},
		{
			name:         "Success with cached block",
			numberOrTag:  "0x7b",
			showDetails:  false,
			mockResponse: expectedBlock,
			expectNil:    false,
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for hex block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x7b").
					Return(int64(123), nil)

				cacheKey := fmt.Sprintf("eth_getBlockByNumber_%d_%t", 123, false)
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
		{
			name:         "Success with show details true",
			numberOrTag:  "0x7b",
			showDetails:  true,
			mockResponse: expectedBlock,
			mockResults: []domain.ContractResults{{
				Hash:             "0xtx1",
				Result:           "SUCCESS",
				BlockHash:        expectedBlock.Hash,
				BlockNumber:      int64(expectedBlock.Number),
				TransactionIndex: 0,
				From:             "0x" + strings.Repeat("2", 40),
				To:               "0x" + strings.Repeat("3", 40),
			}},
			expectNil: false,
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for hex block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x7b").
					Return(int64(123), nil)

				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("eth_getBlockByNumber_%d_%t", 123, true)
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("123").
					Return(expectedBlock)

				mockClient.EXPECT().
					GetContractResults(expectedBlock.Timestamp).
					Return([]domain.ContractResults{{
						Hash:             "0xtx1",
						Result:           "SUCCESS",
						BlockHash:        expectedBlock.Hash,
						BlockNumber:      int64(expectedBlock.Number),
						TransactionIndex: 0,
						From:             "0x" + strings.Repeat("2", 40),
						To:               "0x" + strings.Repeat("3", 40),
					}})

				// Mock resolveEvmAddress for 'from' address
				fromCacheKey := fmt.Sprintf("evm_address_%s", "0x"+strings.Repeat("2", 40))
				cacheService.EXPECT().
					Get(gomock.Any(), fromCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById("0x"+strings.Repeat("2", 40)).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById("0x"+strings.Repeat("2", 40)).
					Return(&domain.AccountResponse{
						EvmAddress: "0x" + strings.Repeat("2", 40),
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), fromCacheKey, "0x"+strings.Repeat("2", 40), service.DefaultExpiration).
					Return(nil)

				// Mock resolveEvmAddress for 'to' address
				toCacheKey := fmt.Sprintf("evm_address_%s", "0x"+strings.Repeat("3", 40))
				cacheService.EXPECT().
					Get(gomock.Any(), toCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				mockClient.EXPECT().
					GetContractById("0x"+strings.Repeat("3", 40)).
					Return(nil, errors.New("not found"))

				mockClient.EXPECT().
					GetAccountById("0x"+strings.Repeat("3", 40)).
					Return(&domain.AccountResponse{
						EvmAddress: "0x" + strings.Repeat("3", 40),
					}, nil)

				cacheService.EXPECT().
					Set(gomock.Any(), toCacheKey, "0x"+strings.Repeat("3", 40), service.DefaultExpiration).
					Return(nil)

				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			s := service.NewEthService(nil, mockClient, commonService, logger, nil, defaultChainId, cacheService)
			result, errRpc := s.GetBlockByNumber(tc.numberOrTag, tc.showDetails)

			if tc.name == "Invalid hex number" {
				assert.NotNil(t, errRpc)
				assert.Equal(t, domain.NewRPCError(domain.ServerError, "Invalid block number"), errRpc)
				return
			}

			if tc.expectNil {
				assert.Nil(t, result)
				assert.Nil(t, errRpc)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, errRpc)

				block, ok := result.(*domain.Block)
				assert.True(t, ok)
				if ok {
					if tc.mockResponse != nil {
						assert.Equal(t, fmt.Sprintf("0x%x", tc.mockResponse.Number), *block.Number)
						assert.Equal(t, tc.mockResponse.Hash, *block.Hash)
						assert.Equal(t, tc.mockResponse.PreviousHash, block.ParentHash)
						assert.Equal(t, fmt.Sprintf("0x%x", tc.mockResponse.GasUsed), block.GasUsed)
						assert.Equal(t, fmt.Sprintf("0x%x", tc.mockResponse.Size), block.Size)
						assert.Equal(t, tc.mockResponse.LogsBloom, block.LogsBloom)
					}

					if !strings.Contains(tc.name, "cached") {
						if tc.showDetails {
							assert.Equal(t, len(tc.mockResults), len(block.Transactions))
							// For show details true, just verify the transaction exists
							assert.NotNil(t, block.Transactions[0])
						} else {
							assert.Equal(t, len(tc.mockResults), len(block.Transactions))
							for i, tx := range tc.mockResults {
								assert.Equal(t, tx.Hash, block.Transactions[i])
							}
						}
					}
				}
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	cacheService := mocks.NewMockCacheService(ctrl)

	mockClient := mocks.NewMockMirrorClient(ctrl)

	s := service.NewEthService(
		nil,
		mockClient,
		nil,
		logger,
		nil,
		defaultChainId,
		cacheService,
	)

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
					GetBalance("0x1234567890123456789012345678901234567890", "0").
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

	// Setup expectations for getting balance with "0" timestamp
	mockClient.EXPECT().
		GetBalance("0x123", "0").
		Return("0x2a")

	s := service.NewEthService(
		nil,
		mockClient,
		nil,
		logger,
		nil,
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
				To: "2023-01-01T00:00:00.000Z",
			},
		})

	// Setup expectations for getting balance
	mockClient.EXPECT().
		GetBalance("0x123", "2023-01-01T00:00:00.000Z").
		Return("0x0")

	s := service.NewEthService(
		nil,
		mockClient,
		nil,
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
		nil,
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
		nil,
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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

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

			result, errRpc := s.Call(tc.transaction, tc.blockParam)

			if tc.expectError {
				assert.NotNil(t, errRpc)
				assert.Equal(t, -32000, errRpc.Code)
			} else {
				assert.Nil(t, errRpc)
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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

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

			result, errRpc := s.EstimateGas(tc.transaction, tc.blockParam)

			if tc.expectError {
				assert.NotNil(t, errRpc)
				assert.Equal(t, -32000, errRpc.Code)
			} else {
				assert.Nil(t, errRpc)
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

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

	// Common test data
	testHash := "0x5d019848d6dad96bc3a9e947350975cd16cf1c51efd4d5b9a273803446fbbb43"
	toAddress := "0x" + strings.Repeat("3", 40)
	fromAddress := "0x" + strings.Repeat("2", 40)
	baseContractResult := domain.ContractResultResponse{
		BlockNumber:        123,
		BlockHash:          "0x" + strings.Repeat("1", 64),
		Hash:               testHash,
		From:               fromAddress,
		To:                 toAddress,
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
				assert.Equal(t, toAddress, *tx.To)
				assert.Equal(t, fromAddress, tx.From)
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
				assert.Equal(t, toAddress, *tx.To)
				assert.Equal(t, fromAddress, tx.From)
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
				assert.Equal(t, toAddress, *tx.To)
				assert.Equal(t, fromAddress, tx.From)
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
			// Set up cache expectations for transaction lookup
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

				// Set up cache expectations for 'to' address resolution
				var cachedToAddr string
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("evm_address_%s", result.To), &cachedToAddr).
					DoAndReturn(func(_ interface{}, _ string, result *string) error {
						*result = toAddress
						return nil
					}).
					Times(1)

				// Set up cache expectations for 'from' address resolution
				var cachedFromAddr string
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("evm_address_%s", result.From), &cachedFromAddr).
					DoAndReturn(func(_ interface{}, _ string, result *string) error {
						*result = fromAddress
						return nil
					}).
					Times(1)

				// Set up cache expectations for storing transaction
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("eth_getTransactionByHash_%s", tc.hash), gomock.Any(), service.DefaultExpiration).
					Return(nil).
					Times(1)
			}

			result, errMap := s.GetTransactionByHash(tc.hash)
			assert.Nil(t, errMap)
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
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)

	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

	txHash := "0x123"
	blockHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	testCases := []struct {
		name        string
		hash        string
		mockResult  domain.ContractResultResponse
		mockBlock   *domain.BlockResponse
		mockFee     int64
		expectError bool
	}{
		{
			name: "successful_transaction_receipt",
			hash: txHash,
			mockResult: domain.ContractResultResponse{
				BlockHash:          blockHash,
				BlockNumber:        123,
				BlockGasUsed:       150000,
				GasUsed:            100000,
				From:               "0xabc",
				To:                 "0xdef",
				TransactionIndex:   1,
				Status:             "0x1",
				Type:               nil,
				Logs:               []domain.MirroNodeLogs{},
				Bloom:              "0x0",
				Address:            "0x0",
				FunctionParameters: "0000000000000000000000000000000000000000000000000000000000000000",
				CallResult:         "",
			},
			mockBlock: &domain.BlockResponse{
				Hash: blockHash,
				Timestamp: domain.Timestamp{
					From: "123",
					To:   "456",
				},
			},
			mockFee:     1000000000,
			expectError: false,
		},
		{
			name:        "transaction_not_found",
			hash:        "0xnonexistent",
			mockResult:  domain.ContractResultResponse{},
			mockBlock:   nil,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up cache expectations for transaction receipt
			cacheService.EXPECT().
				Get(gomock.Any(), fmt.Sprintf("%s_%s", service.GetTransactionReceipt, tc.hash), gomock.Any()).
				Return(errors.New("not found")).
				Times(1)

			// Set up mock expectations for GetContractResult
			if tc.hash == "0xnonexistent" {
				mockClient.EXPECT().
					GetContractResult(tc.hash).
					Return(nil).
					Times(1)
			} else {
				mockClient.EXPECT().
					GetContractResult(tc.hash).
					Return(tc.mockResult).
					Times(1)

				// Mock address resolution for 'from' address
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("evm_address_%s", tc.mockResult.From), gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				// Mock concurrent address resolution calls for 'from' address
				mockClient.EXPECT().
					GetContractById(tc.mockResult.From).
					Return(&domain.ContractResponse{
						EvmAddress: tc.mockResult.From,
					}, nil).
					AnyTimes()

				mockClient.EXPECT().
					GetAccountById(tc.mockResult.From).
					Return(&domain.AccountResponse{
						EvmAddress: tc.mockResult.From,
					}, nil).
					AnyTimes()

				// Mock token check for 'from' address
				if strings.HasPrefix(tc.mockResult.From, "0x000000000000") {
					mockClient.EXPECT().
						GetTokenById(gomock.Any()).
						Return(&domain.TokenResponse{}, nil).
						AnyTimes()
				}

				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("evm_address_%s", tc.mockResult.From), tc.mockResult.From, service.DefaultExpiration).
					Return(nil).
					Times(1)

				// Mock address resolution for 'to' address
				cacheService.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("evm_address_%s", tc.mockResult.To), gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				// Mock concurrent address resolution calls for 'to' address
				mockClient.EXPECT().
					GetContractById(tc.mockResult.To).
					Return(&domain.ContractResponse{
						EvmAddress: tc.mockResult.To,
					}, nil).
					AnyTimes()

				mockClient.EXPECT().
					GetAccountById(tc.mockResult.To).
					Return(&domain.AccountResponse{
						EvmAddress: tc.mockResult.To,
					}, nil).
					AnyTimes()

				// Mock token check for 'to' address
				if strings.HasPrefix(tc.mockResult.To, "0x000000000000") {
					mockClient.EXPECT().
						GetTokenById(gomock.Any()).
						Return(&domain.TokenResponse{}, nil).
						AnyTimes()
				}

				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("evm_address_%s", tc.mockResult.To), tc.mockResult.To, service.DefaultExpiration).
					Return(nil).
					Times(1)

				// Mock GetBlockByHashOrNumber for gas price
				mockClient.EXPECT().
					GetBlockByHashOrNumber(tc.mockResult.BlockHash[:66]).
					Return(tc.mockBlock).
					Times(1)

				// Mock GetNetworkFees
				mockClient.EXPECT().
					GetNetworkFees(tc.mockBlock.Timestamp.From, "").
					Return(tc.mockFee, nil).
					Times(1)

				// Mock cache Set for receipt
				cacheService.EXPECT().
					Set(gomock.Any(), fmt.Sprintf("%s_%s", service.GetTransactionReceipt, tc.hash), gomock.Any(), service.DefaultExpiration).
					Return(nil).
					Times(1)
			}

			result, errMap := s.GetTransactionReceipt(tc.hash)
			if tc.expectError {
				assert.NotNil(t, errMap)
			} else {
				assert.Nil(t, errMap)
				if tc.hash == "0xnonexistent" {
					assert.Nil(t, result)
				} else {
					receipt, ok := result.(domain.TransactionReceipt)
					assert.True(t, ok)
					assert.Equal(t, tc.mockResult.BlockHash[:66], receipt.BlockHash)
					assert.Equal(t, "0x7b", receipt.BlockNumber) // 123 in hex
					assert.Equal(t, tc.mockResult.From, receipt.From)
					assert.Equal(t, tc.mockResult.To, receipt.To)
					assert.Equal(t, "0x249f0", receipt.CumulativeGasUsed) // 150000 in hex
					assert.Equal(t, "0x186a0", receipt.GasUsed)           // 100000 in hex
					assert.Equal(t, "0x1", receipt.Status)
					assert.Equal(t, tc.hash, receipt.TransactionHash)
					assert.Equal(t, "0x1", receipt.TransactionIndex)
				}
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
	commonService := mocks.NewMockCommonService(ctrl)

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
				// Mock cache get attempt for block number
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_blockNumber", gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				// Mock GetBlockNumber
				commonService.EXPECT().
					GetBlockNumber().
					Return(interface{}("0x64"), nil).
					Times(1)

				// Mock cache set for block number
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_blockNumber", "0x64", service.ShortExpiration).
					Return(nil).
					Times(1)

				// Mock GetBlockNumberByNumberOrTag
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil).
					Times(1)

				// Mock gas price retrieval
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
				// Mock cache get attempt for block number
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_blockNumber", gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				// Mock GetBlockNumber from commonService
				commonService.EXPECT().
					GetBlockNumber().
					Return("0x64", nil). // 100 in hex
					Times(1)

				// Mock cache set for block number
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_blockNumber", "0x64", service.ShortExpiration).
					Return(nil).
					Times(1)

				// Mock GetBlockNumberByNumberOrTag
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil).
					Times(1)

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
				// Mock cache get attempt for block number
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_blockNumber", gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				// Mock GetBlockNumber from commonService
				commonService.EXPECT().
					GetBlockNumber().
					Return("0x64", nil). // 100 in hex
					Times(1)

				// Mock cache set for block number
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_blockNumber", "0x64", service.ShortExpiration).
					Return(nil).
					Times(1)

				// Mock GetBlockNumberByNumberOrTag
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil).
					Times(1)
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
				// Mock cache get attempt for block number
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_blockNumber", gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				// Mock GetBlockNumber from commonService
				commonService.EXPECT().
					GetBlockNumber().
					Return("0x64", nil). // 100 in hex
					Times(1)

				// Mock cache set for block number
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_blockNumber", "0x64", service.ShortExpiration).
					Return(nil).
					Times(1)

				// Mock GetBlockNumberByNumberOrTag to fail with proper RPCError
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0xinvalid").
					Return(int64(0), domain.NewRPCError(domain.ServerError, "Invalid block number")).
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
			setupMocks: func() { // Mock cache get attempt for block number
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_blockNumber", gomock.Any()).
					Return(errors.New("not found")).
					Times(1)

				// Mock GetBlockNumber from commonService
				commonService.EXPECT().
					GetBlockNumber().
					Return(nil, nil). // 100 in hex
					Times(1)

				// Mock cache set for block number
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_blockNumber", nil, service.ShortExpiration).
					Return(nil).
					Times(1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := service.NewEthService(nil, mockClient, commonService, logger, nil, "0x12a", cacheService)

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
	commonService := mocks.NewMockCommonService(ctrl)

	s := service.NewEthService(nil, mockClient, commonService, logger, nil, defaultChainId, cacheService)

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
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil)

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
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("earliest").
					Return(int64(0), nil)

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
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x50").
					Return(int64(80), nil)

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
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x999").
					Return(int64(2457), nil)

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
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil)

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
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil)

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

			result, errRpc := s.GetStorageAt(tc.address, tc.slot, tc.blockParam)

			if tc.expectError {
				assert.NotNil(t, errRpc)
				assert.Equal(t, -32000, errRpc.Code)
			} else {
				assert.Nil(t, errRpc)
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
	commonService := mocks.NewMockCommonService(ctrl)

	s := service.NewEthService(nil, mockClient, commonService, logger, nil, defaultChainId, cacheService)

	testCases := []struct {
		name           string
		logParams      domain.LogParams
		setupMocks     func()
		expectedResult interface{}
		expectError    bool
		expectedCode   int
	}{
		{
			name: "Success with block hash",
			logParams: domain.LogParams{
				BlockHash: "0x123abc",
				Address:   []string{"0x742d35Cc6634C0532925a3b844Bc454e4438f44e"},
				Topics:    []string{"0xtopic1", "0xtopic2"},
			},
			setupMocks: func() {
				commonService.EXPECT().
					GetLogs(domain.LogParams{
						BlockHash: "0x123abc",
						Address:   []string{"0x742d35Cc6634C0532925a3b844Bc454e4438f44e"},
						Topics:    []string{"0xtopic1", "0xtopic2"},
					}).
					Return([]domain.Log{
						{
							Address:          "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
							BlockHash:        "0x123abc",
							BlockNumber:      "0x64", // 100 in hex
							Data:             "0xdata",
							LogIndex:         "0x0",
							Removed:          false,
							Topics:           []string{"0xtopic1", "0xtopic2"},
							TransactionHash:  "0xtxhash",
							TransactionIndex: "0x1",
						},
					}, nil)
			},
			expectedResult: []domain.Log{
				{
					Address:          "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
					BlockHash:        "0x123abc",
					BlockNumber:      "0x64", // 100 in hex
					Data:             "0xdata",
					LogIndex:         "0x0",
					Removed:          false,
					Topics:           []string{"0xtopic1", "0xtopic2"},
					TransactionHash:  "0xtxhash",
					TransactionIndex: "0x1",
				},
			},
			expectError:  false,
			expectedCode: 0,
		},
		{
			name: "Invalid block hash",
			logParams: domain.LogParams{
				BlockHash: "0xinvalid",
			},
			setupMocks: func() {
				commonService.EXPECT().
					GetLogs(domain.LogParams{
						BlockHash: "0xinvalid",
					}).
					Return([]domain.Log{}, nil)
			},
			expectedResult: []domain.Log{},
			expectError:    false,
			expectedCode:   0,
		},
		{
			name: "Latest block fetch failure",
			logParams: domain.LogParams{
				FromBlock: "0x1",
				ToBlock:   "latest",
			},
			setupMocks: func() {
				commonService.EXPECT().
					GetLogs(domain.LogParams{
						FromBlock: "0x1",
						ToBlock:   "latest",
					}).
					Return(nil, domain.NewRPCError(domain.ServerError, "Failed to fetch latest block"))
			},
			expectedResult: nil,
			expectError:    true,
			expectedCode:   -32000,
		},
		{
			name: "From block greater than to block",
			logParams: domain.LogParams{
				FromBlock: "0x10",
				ToBlock:   "0x5",
			},
			setupMocks: func() {
				commonService.EXPECT().
					GetLogs(domain.LogParams{
						FromBlock: "0x10",
						ToBlock:   "0x5",
					}).
					Return(nil, domain.NewRPCError(domain.InvalidParams, "FromBlock is greater than ToBlock"))
			},
			expectedResult: nil,
			expectError:    true,
			expectedCode:   -32602,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			result, errRpc := s.GetLogs(tc.logParams)

			if tc.expectError {
				assert.NotNil(t, errRpc)
				if tc.expectedCode != 0 {
					assert.Equal(t, tc.expectedCode, errRpc.Code)
				}
			} else {
				assert.Nil(t, errRpc)
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
	commonService := mocks.NewMockCommonService(ctrl)

	s := service.NewEthService(nil, mockClient, commonService, logger, nil, defaultChainId, cacheService)

	testCases := []struct {
		name            string
		blockParam      string
		mockLatestBlock map[string]interface{}
		mockResponse    *domain.BlockResponse
		expectedResult  interface{}
		expectedError   *domain.RPCError
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
				// Mock GetBlockNumberByNumberOrTag
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x7b").
					Return(int64(123), nil)

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
					Set(gomock.Any(), "eth_getBlockTransactionCountByNumber_123", "0x5", service.DefaultExpiration).
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
				// Mock GetBlockNumberByNumberOrTag
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil)

				// Mock cache get attempt
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_getBlockTransactionCountByNumber_100", gomock.Any()).
					Return(fmt.Errorf("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("100").
					Return(&domain.BlockResponse{Count: 10})

				// Mock cache set with the result
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_getBlockTransactionCountByNumber_100", "0xa", service.DefaultExpiration).
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
				// Mock GetBlockNumberByNumberOrTag
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("earliest").
					Return(int64(0), nil)

				// Mock cache get attempt
				cacheService.EXPECT().
					Get(gomock.Any(), "eth_getBlockTransactionCountByNumber_0", gomock.Any()).
					Return(fmt.Errorf("not found"))

				mockClient.EXPECT().
					GetBlockByHashOrNumber("0").
					Return(&domain.BlockResponse{Count: 1})

				// Mock cache set with the result
				cacheService.EXPECT().
					Set(gomock.Any(), "eth_getBlockTransactionCountByNumber_0", "0x1", service.DefaultExpiration).
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
				// Mock GetBlockNumberByNumberOrTag
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x999").
					Return(int64(2457), nil)

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
			expectedError:  domain.NewRPCError(domain.ServerError, "Invalid block number"),
			setupMock: func() {
				// Mock GetBlockNumberByNumberOrTag to return error
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0xinvalid").
					Return(int64(0), domain.NewRPCError(domain.ServerError, "Invalid block number"))
			},
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
	s := service.NewEthService(nil, mockClient, nil, logger, nil, defaultChainId, cacheService)

	baseContractResult := domain.ContractResults{
		BlockNumber:      123,
		BlockHash:        "0x" + strings.Repeat("1", 64),
		Hash:             "0x" + strings.Repeat("a", 64),
		From:             "0x" + strings.Repeat("2", 40),
		To:               "0x" + strings.Repeat("3", 40),
		GasUsed:          21000,
		GasPrice:         "0x5678",
		TransactionIndex: 1,
		Amount:           1000000,
		V:                27,
		R:                "0x" + strings.Repeat("4", 64),
		S:                "0x" + strings.Repeat("5", 64),
		Nonce:            5,
		Type:             0,
		ChainID:          "0x1",
	}

	blockHash := "0x" + strings.Repeat("1", 64)

	testCases := []struct {
		name           string
		blockHash      string
		index          string
		mockResult     *domain.ContractResults
		expectedResult interface{}
		expectedError  *domain.RPCError
		setupMocks     func()
		checkFields    func(t *testing.T, result interface{})
	}{
		{
			name:      "successful transaction retrieval",
			blockHash: blockHash,
			index:     "0x1",
			setupMocks: func() {
				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("%s_%s_%s", service.GetTransactionByBlockHashAndIndex, blockHash, "0x1")
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("cache miss"))

				// Mock getting contract result
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(&baseContractResult, nil)

				// Mock cache service for from address
				fromCacheKey := fmt.Sprintf("evm_address_%s", baseContractResult.From)
				cacheService.EXPECT().
					Get(gomock.Any(), fromCacheKey, gomock.Any()).
					DoAndReturn(func(_ interface{}, _ string, result *string) error {
						*result = baseContractResult.From
						return nil
					})

				// Mock cache service for to address
				toCacheKey := fmt.Sprintf("evm_address_%s", baseContractResult.To)
				cacheService.EXPECT().
					Get(gomock.Any(), toCacheKey, gomock.Any()).
					DoAndReturn(func(_ interface{}, _ string, result *string) error {
						*result = baseContractResult.To
						return nil
					})

				// Mock cache set for transaction
				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction)
				assert.True(t, ok)
				assert.Equal(t, "0x7b", *tx.BlockNumber)
				assert.Equal(t, baseContractResult.From, tx.From)
				assert.Equal(t, baseContractResult.To, *tx.To)
				assert.Equal(t, "0x5208", tx.Gas)              // 21000 in hex
				assert.Equal(t, "0xc953642ae000", tx.GasPrice) // 0x5678 * 10^10
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x5", tx.Nonce) // 5 in hex
				assert.Equal(t, "0x1", *tx.TransactionIndex)
				assert.Equal(t, "0xf4240", tx.Value) // 1000000 in hex
				assert.Equal(t, "0x1b", tx.V)        // 27 in hex
				assert.Equal(t, baseContractResult.R, tx.R)
				assert.Equal(t, baseContractResult.S, tx.S)
				assert.Equal(t, "0x0", tx.Type)
				assert.Equal(t, "0x1", *tx.ChainId)
			},
		},
		{
			name:      "invalid transaction index",
			blockHash: blockHash,
			index:     "invalid",
			setupMocks: func() {
				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("%s_%s_%s", service.GetTransactionByBlockHashAndIndex, blockHash, "invalid")
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("cache miss"))
			},
			expectedError: domain.NewRPCError(domain.ServerError, "Failed to parse hex value"),
		},
		{
			name:      "transaction not found",
			blockHash: blockHash,
			index:     "0x1",
			setupMocks: func() {
				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("%s_%s_%s", service.GetTransactionByBlockHashAndIndex, blockHash, "0x1")
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("cache miss"))

				// Mock getting contract result
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(nil, nil)

				// Mock cache set for nil result
				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, nil, service.DefaultExpiration).
					Return(nil)
			},
			expectedResult: nil,
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
	commonService := mocks.NewMockCommonService(ctrl)
	s := service.NewEthService(nil, mockClient, commonService, logger, nil, defaultChainId, cacheService)

	baseContractResult := domain.ContractResults{
		BlockNumber:      123,
		BlockHash:        "0x" + strings.Repeat("1", 64),
		Hash:             "0x" + strings.Repeat("a", 64),
		From:             "0x" + strings.Repeat("2", 40),
		To:               "0x" + strings.Repeat("3", 40),
		GasUsed:          21000,
		GasPrice:         "0x5678",
		TransactionIndex: 1,
		Amount:           1000000,
		V:                27,
		R:                "0x" + strings.Repeat("4", 64),
		S:                "0x" + strings.Repeat("5", 64),
		Nonce:            5,
		Type:             0,
		ChainID:          "0x1",
	}

	testCases := []struct {
		name           string
		blockNumber    string
		index          string
		mockResult     *domain.ContractResults
		expectedResult interface{}
		expectedError  *domain.RPCError
		setupMocks     func()
		checkFields    func(t *testing.T, result interface{})
	}{
		{
			name:        "successful transaction retrieval with latest block",
			blockNumber: "latest",
			index:       "0x1",
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for latest block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(123), nil)

				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("%s_%d_%s", service.GetTransactionByBlockNumberAndIndex, 123, "0x1")
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("cache miss"))

				// Mock getting contract result
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(&baseContractResult, nil)

				// Mock cache service for from address
				fromCacheKey := fmt.Sprintf("evm_address_%s", baseContractResult.From)
				cacheService.EXPECT().
					Get(gomock.Any(), fromCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock concurrent resolution calls for from address
				mockClient.EXPECT().
					GetContractById(baseContractResult.From).
					Return(&domain.ContractResponse{
						EvmAddress: baseContractResult.From,
					}, nil).
					AnyTimes()

				mockClient.EXPECT().
					GetAccountById(baseContractResult.From).
					Return(&domain.AccountResponse{
						EvmAddress: baseContractResult.From,
					}, nil).
					AnyTimes()

				// Mock cache set for from address
				cacheService.EXPECT().
					Set(gomock.Any(), fromCacheKey, baseContractResult.From, service.DefaultExpiration).
					Return(nil)

				// Mock cache service for to address
				toCacheKey := fmt.Sprintf("evm_address_%s", baseContractResult.To)
				cacheService.EXPECT().
					Get(gomock.Any(), toCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock concurrent resolution calls for to address
				mockClient.EXPECT().
					GetContractById(baseContractResult.To).
					Return(&domain.ContractResponse{
						EvmAddress: baseContractResult.To,
					}, nil).
					AnyTimes()

				mockClient.EXPECT().
					GetAccountById(baseContractResult.To).
					Return(&domain.AccountResponse{
						EvmAddress: baseContractResult.To,
					}, nil).
					AnyTimes()

				// Mock cache set for to address
				cacheService.EXPECT().
					Set(gomock.Any(), toCacheKey, baseContractResult.To, service.DefaultExpiration).
					Return(nil)

				// Mock cache set for transaction
				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction)
				assert.True(t, ok)
				assert.Equal(t, "0x7b", *tx.BlockNumber)
				assert.Equal(t, baseContractResult.From, tx.From)
				assert.Equal(t, baseContractResult.To, *tx.To)
				assert.Equal(t, "0x5208", tx.Gas)              // 21000 in hex
				assert.Equal(t, "0xc953642ae000", tx.GasPrice) // 0x5678 * 10^10
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x5", tx.Nonce) // 5 in hex
				assert.Equal(t, "0x1", *tx.TransactionIndex)
				assert.Equal(t, "0xf4240", tx.Value) // 1000000 in hex
				assert.Equal(t, "0x1b", tx.V)        // 27 in hex
				assert.Equal(t, baseContractResult.R, tx.R)
				assert.Equal(t, baseContractResult.S, tx.S)
				assert.Equal(t, "0x0", tx.Type)
				assert.Equal(t, "0x1", *tx.ChainId)
			},
		},
		{
			name:        "successful transaction retrieval with hex block",
			blockNumber: "0x7b", // 123 in hex
			index:       "0x1",
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for hex block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x7b").
					Return(int64(123), nil)

				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("%s_%d_%s", service.GetTransactionByBlockNumberAndIndex, 123, "0x1")
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("cache miss"))

				// Mock getting contract result
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(&baseContractResult, nil)

				// Mock cache service for from address
				fromCacheKey := fmt.Sprintf("evm_address_%s", baseContractResult.From)
				cacheService.EXPECT().
					Get(gomock.Any(), fromCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock concurrent resolution calls for from address
				mockClient.EXPECT().
					GetContractById(baseContractResult.From).
					Return(&domain.ContractResponse{
						EvmAddress: baseContractResult.From,
					}, nil).
					AnyTimes()

				mockClient.EXPECT().
					GetAccountById(baseContractResult.From).
					Return(&domain.AccountResponse{
						EvmAddress: baseContractResult.From,
					}, nil).
					AnyTimes()

				// Mock cache set for from address
				cacheService.EXPECT().
					Set(gomock.Any(), fromCacheKey, baseContractResult.From, service.DefaultExpiration).
					Return(nil)

				// Mock cache service for to address
				toCacheKey := fmt.Sprintf("evm_address_%s", baseContractResult.To)
				cacheService.EXPECT().
					Get(gomock.Any(), toCacheKey, gomock.Any()).
					Return(errors.New("not found"))

				// Mock concurrent resolution calls for to address
				mockClient.EXPECT().
					GetContractById(baseContractResult.To).
					Return(&domain.ContractResponse{
						EvmAddress: baseContractResult.To,
					}, nil).
					AnyTimes()

				mockClient.EXPECT().
					GetAccountById(baseContractResult.To).
					Return(&domain.AccountResponse{
						EvmAddress: baseContractResult.To,
					}, nil).
					AnyTimes()

				// Mock cache set for to address
				cacheService.EXPECT().
					Set(gomock.Any(), toCacheKey, baseContractResult.To, service.DefaultExpiration).
					Return(nil)

				// Mock cache set for transaction
				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, gomock.Any(), service.DefaultExpiration).
					Return(nil)
			},
			checkFields: func(t *testing.T, result interface{}) {
				tx, ok := result.(domain.Transaction)
				assert.True(t, ok)
				assert.Equal(t, "0x7b", *tx.BlockNumber)
				assert.Equal(t, baseContractResult.From, tx.From)
				assert.Equal(t, baseContractResult.To, *tx.To)
				assert.Equal(t, "0x5208", tx.Gas)              // 21000 in hex
				assert.Equal(t, "0xc953642ae000", tx.GasPrice) // 0x5678 * 10^10
				assert.Equal(t, baseContractResult.Hash, tx.Hash)
				assert.Equal(t, "0x5", tx.Nonce) // 5 in hex
				assert.Equal(t, "0x1", *tx.TransactionIndex)
				assert.Equal(t, "0xf4240", tx.Value) // 1000000 in hex
				assert.Equal(t, "0x1b", tx.V)        // 27 in hex
				assert.Equal(t, baseContractResult.R, tx.R)
				assert.Equal(t, baseContractResult.S, tx.S)
				assert.Equal(t, "0x0", tx.Type)
				assert.Equal(t, "0x1", *tx.ChainId)
			},
		},
		{
			name:        "invalid block number",
			blockNumber: "invalid",
			index:       "0x1",
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag to return error for invalid block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("invalid").
					Return(int64(0), domain.NewRPCError(domain.ServerError, "Invalid block number"))
			},
			expectedError: domain.NewRPCError(domain.ServerError, "Invalid block number"),
		},
		{
			name:        "invalid transaction index",
			blockNumber: "0x7b",
			index:       "invalid",
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for hex block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x7b").
					Return(int64(123), nil)

				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("%s_%d_%s", service.GetTransactionByBlockNumberAndIndex, 123, "invalid")
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("cache miss"))
			},
			expectedError: domain.NewRPCError(domain.ServerError, "Failed to parse hex value"),
		},
		{
			name:        "transaction not found",
			blockNumber: "0x7b",
			index:       "0x1",
			setupMocks: func() {
				// Mock GetBlockNumberByNumberOrTag for hex block
				commonService.EXPECT().
					GetBlockNumberByNumberOrTag("0x7b").
					Return(int64(123), nil)

				// Mock cache miss for transaction
				cacheKey := fmt.Sprintf("%s_%d_%s", service.GetTransactionByBlockNumberAndIndex, 123, "0x1")
				cacheService.EXPECT().
					Get(gomock.Any(), cacheKey, gomock.Any()).
					Return(errors.New("cache miss"))

				// Mock getting contract result
				mockClient.EXPECT().
					GetContractResultWithRetry(gomock.Any()).
					Return(nil, nil)

				// Mock cache set for nil result
				cacheService.EXPECT().
					Set(gomock.Any(), cacheKey, nil, service.DefaultExpiration).
					Return(nil)
			},
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			result, errRpc := s.GetTransactionByBlockNumberAndIndex(tc.blockNumber, tc.index)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, errRpc)
			} else {
				assert.Nil(t, errRpc)
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
		nil,
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

		// Set up concurrent resolution expectations
		mockClient.EXPECT().
			GetContractById(address).
			Return(nil, errors.New("not found")).
			AnyTimes()

		mockClient.EXPECT().
			GetAccountById(address).
			Return(nil, errors.New("not found")).
			AnyTimes()

		mockClient.EXPECT().
			GetTokenById(gomock.Any()).
			Return(nil, errors.New("not a token")).
			AnyTimes()

		// Expect GetContractByteCode call
		mockHederaClient.EXPECT().
			GetContractByteCode(int64(0), int64(0), address).
			Return([]byte{0x60, 0x60, 0x60}, nil).
			Times(1)

		cacheService.EXPECT().
			Set(gomock.Any(), cacheKey, runtimeBytecode, DefaultExpiration).
			Return(nil)

		result, errMap := s.GetCode(address, blockNumber)

		assert.Equal(t, runtimeBytecode, result)
		assert.Nil(t, errMap)
	})

	t.Run("Fallback to Hedera client", func(t *testing.T) {
		address := "0x456"
		blockNumber := "latest"
		bytecode := []byte{1, 2, 3}
		expectedResponse := fmt.Sprintf("0x%x", bytecode)

		cacheKey := fmt.Sprintf("%s_%s_%s", GetCode, address, blockNumber)

		// First expect cache check
		cacheService.EXPECT().
			Get(gomock.Any(), cacheKey, gomock.Any()).
			Return(errors.New("not found"))

		// Then expect concurrent resolution attempts
		mockClient.EXPECT().
			GetContractById(address).
			Return(nil, fmt.Errorf("not found")).
			AnyTimes()

		mockClient.EXPECT().
			GetAccountById(address).
			Return(nil, fmt.Errorf("not found")).
			AnyTimes()

		mockClient.EXPECT().
			GetTokenById(gomock.Any()).
			Return(nil, fmt.Errorf("not a token")).
			AnyTimes()

		// Then expect Hedera client call with exact parameters
		mockHederaClient.EXPECT().
			GetContractByteCode(int64(0), int64(0), address).
			Return(bytecode, nil).
			Times(1)

		// Finally expect cache set with exact parameters
		cacheService.EXPECT().
			Set(gomock.Any(), cacheKey, expectedResponse, service.DefaultExpiration).
			Return(nil).
			Times(1)

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
		// First expect cache check
		cacheService.EXPECT().
			Get(gomock.Any(), cacheKey, gomock.Any()).
			Return(errors.New("not found"))

		// Then expect concurrent resolution attempts
		mockClient.EXPECT().
			GetContractById(address).
			Return(nil, fmt.Errorf("not found")).
			AnyTimes()

		mockClient.EXPECT().
			GetAccountById(address).
			Return(nil, fmt.Errorf("not found")).
			AnyTimes()

		// Add token resolution expectation
		mockClient.EXPECT().
			GetTokenById(gomock.Any()).
			Return(nil, fmt.Errorf("not a token")).
			AnyTimes()

		// Finally expect Hedera client call with exact parameters
		mockHederaClient.EXPECT().
			GetContractByteCode(int64(0), int64(0), address).
			Return(nil, fmt.Errorf("hedera client error"))

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
	ethService := service.NewEthService(mockHederaClient, mockMirrorClient, nil, logger, nil, "0x128", mockCacheService)

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
		result, errRpc := ethService.SendRawTransaction("")

		assert.NotNil(t, errRpc)
		assert.Nil(t, result)
		assert.Equal(t, domain.NewRPCError(domain.ServerError, "Failed to parse transaction"), errRpc)
	})
}
