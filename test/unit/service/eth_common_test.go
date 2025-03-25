package service_test

import (
	"fmt"
	"testing"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/service"
	"github.com/LimeChain/Hederium/test/unit/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupCommonTest(t *testing.T) (*gomock.Controller, *mocks.MockMirrorClient, *mocks.MockCacheService, service.CommonService) {
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	mockCache := mocks.NewMockCacheService(ctrl)
	commonService := service.NewCommonService(mockClient, logger, mockCache)

	return ctrl, mockClient, mockCache, commonService
}

func TestGetBlockNumberByNumberOrTag(t *testing.T) {
	ctrl, mockClient, _, commonService := setupCommonTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		input          string
		mockSetup      func()
		expectedResult int64
		expectError    bool
	}{
		{
			name:  "Latest tag",
			input: "latest",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(123)}, nil)
			},
			expectedResult: 123,
			expectError:    false,
		},
		{
			name:  "Pending tag",
			input: "pending",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(123)}, nil)
			},
			expectedResult: 123,
			expectError:    false,
		},
		{
			name:           "Earliest tag",
			input:          "earliest",
			mockSetup:      func() {},
			expectedResult: 0,
			expectError:    false,
		},
		{
			name:           "Hex number",
			input:          "0x7b", // 123 in hex
			mockSetup:      func() {},
			expectedResult: 123,
			expectError:    false,
		},
		{
			name:           "Invalid hex",
			input:          "0xinvalid",
			mockSetup:      func() {},
			expectedResult: 0,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, errRpc := commonService.GetBlockNumberByNumberOrTag(tc.input)

			if tc.expectError {
				assert.NotNil(t, errRpc)
			} else {
				assert.Nil(t, errRpc)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestValidateBlockRange(t *testing.T) {
	ctrl, mockClient, _, commonService := setupCommonTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name        string
		fromBlock   string
		toBlock     string
		mockSetup   func()
		expectError bool
	}{
		{
			name:      "Valid range",
			fromBlock: "0x1",
			toBlock:   "0x2",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)
			},
			expectError: false,
		},
		{
			name:      "From block greater than to block",
			fromBlock: "0x2",
			toBlock:   "0x1",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)
			},
			expectError: true,
		},
		{
			name:      "Missing from block with explicit to block",
			fromBlock: "",
			toBlock:   "0x5",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)
			},
			expectError: true,
		},
		{
			name:      "Latest blocks",
			fromBlock: "latest",
			toBlock:   "latest",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			errRpc := commonService.ValidateBlockRange(tc.fromBlock, tc.toBlock)

			if tc.expectError {
				assert.NotNil(t, errRpc)
			} else {
				assert.Nil(t, errRpc)
			}
		})
	}
}

func TestGetLogsWithParams(t *testing.T) {
	ctrl, mockClient, _, commonService := setupCommonTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		addresses      []string
		params         map[string]interface{}
		mockSetup      func()
		expectedResult []domain.Log
		expectError    bool
	}{
		{
			name:      "Success with no addresses",
			addresses: nil,
			params: map[string]interface{}{
				"timestamp": "gte:1672531200&timestamp=lte:1672531202",
			},
			mockSetup: func() {
				mockClient.EXPECT().
					GetContractResultsLogsWithRetry(map[string]interface{}{
						"timestamp": "gte:1672531200&timestamp=lte:1672531202",
					}).
					Return([]domain.LogEntry{
						{
							Address:          "0xaddress",
							BlockHash:        "0xblockhash",
							BlockNumber:      ptr(int64(1)),
							Data:             "0xdata",
							TransactionHash:  "0xtxhash",
							TransactionIndex: ptr(0),
							Index:            ptr(0),
							Topics:           []string{},
						},
					}, nil)
			},
			expectedResult: []domain.Log{
				{
					Address:          "0xaddress",
					BlockHash:        "0xblockhash",
					BlockNumber:      "0x1",
					Data:             "0xdata",
					LogIndex:         "0x0",
					Removed:          false,
					Topics:           []string{},
					TransactionHash:  "0xtxhash",
					TransactionIndex: "0x0",
				},
			},
			expectError: false,
		},
		{
			name:      "Success with specific address",
			addresses: []string{"0xaddress"},
			params: map[string]interface{}{
				"timestamp": "gte:1672531200&timestamp=lte:1672531202",
			},
			mockSetup: func() {
				mockClient.EXPECT().
					GetContractResultsLogsByAddress("0xaddress", map[string]interface{}{
						"timestamp": "gte:1672531200&timestamp=lte:1672531202",
					}).
					Return([]domain.LogEntry{
						{
							Address:          "0xaddress",
							BlockHash:        "0xblockhash",
							BlockNumber:      ptr(int64(1)),
							Data:             "0xdata",
							TransactionHash:  "0xtxhash",
							TransactionIndex: ptr(0),
							Index:            ptr(0),
							Topics:           []string{},
						},
					}, nil)
			},
			expectedResult: []domain.Log{
				{
					Address:          "0xaddress",
					BlockHash:        "0xblockhash",
					BlockNumber:      "0x1",
					Data:             "0xdata",
					LogIndex:         "0x0",
					Removed:          false,
					Topics:           []string{},
					TransactionHash:  "0xtxhash",
					TransactionIndex: "0x0",
				},
			},
			expectError: false,
		},
		{
			name:      "Error fetching logs",
			addresses: []string{"0xaddress"},
			params: map[string]interface{}{
				"timestamp": "gte:1672531200&timestamp=lte:1672531202",
			},
			mockSetup: func() {
				mockClient.EXPECT().
					GetContractResultsLogsByAddress("0xaddress", map[string]interface{}{
						"timestamp": "gte:1672531200&timestamp=lte:1672531202",
					}).
					Return(nil, fmt.Errorf("failed to fetch logs"))
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, err := commonService.GetLogsWithParams(tc.addresses, tc.params)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestValidateBlockHashAndAddTimestampToParams(t *testing.T) {
	ctrl, mockClient, _, commonService := setupCommonTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		blockHash      string
		params         map[string]interface{}
		mockSetup      func()
		expectError    bool
		expectedParams map[string]interface{}
	}{
		{
			name:      "Valid block hash",
			blockHash: "0x123abc",
			params:    make(map[string]interface{}),
			mockSetup: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0x123abc").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							From: "2023-01-01T00:00:00.000Z",
							To:   "2023-01-01T00:00:01.000Z",
						},
					})
			},
			expectError: false,
			expectedParams: map[string]interface{}{
				"timestamp": "gte:2023-01-01T00:00:00.000Z&timestamp=lte:2023-01-01T00:00:01.000Z",
			},
		},
		{
			name:      "Invalid block hash",
			blockHash: "0xinvalid",
			params:    make(map[string]interface{}),
			mockSetup: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0xinvalid").
					Return(nil)
			},
			expectError:    true,
			expectedParams: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			err := commonService.ValidateBlockHashAndAddTimestampToParams(tc.params, tc.blockHash)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedParams, tc.params)
			}
		})
	}
}

func TestValidateBlockRangeAndAddTimestampToParams(t *testing.T) {
	ctrl, mockClient, _, commonService := setupCommonTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		fromBlock      string
		toBlock        string
		address        []string
		params         map[string]interface{}
		mockSetup      func()
		expectOk       bool
		expectError    bool
		expectedParams map[string]interface{}
	}{
		{
			name:      "Valid block range",
			fromBlock: "0x1",
			toBlock:   "0x2",
			address:   []string{"0xaddress"},
			params:    make(map[string]interface{}),
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)

				mockClient.EXPECT().
					GetBlockByHashOrNumber("1").
					Return(&domain.BlockResponse{
						Number: 1,
						Timestamp: domain.Timestamp{
							From: "1672531200",
							To:   "1672531201",
						},
					})

				mockClient.EXPECT().
					GetBlockByHashOrNumber("2").
					Return(&domain.BlockResponse{
						Number: 2,
						Timestamp: domain.Timestamp{
							From: "1672531201",
							To:   "1672531202",
						},
					})
			},
			expectOk:    true,
			expectError: false,
			expectedParams: map[string]interface{}{
				"timestamp": "gte:1672531200&timestamp=lte:1672531202",
			},
		},
		{
			name:      "Block range too large",
			fromBlock: "0x1",
			toBlock:   "0x64", // 100 in hex
			address:   []string{"0xaddress"},
			params:    make(map[string]interface{}),
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)

				mockClient.EXPECT().
					GetBlockByHashOrNumber("1").
					Return(&domain.BlockResponse{
						Number: 1,
						Timestamp: domain.Timestamp{
							From: "1672531200",
							To:   "1672531201",
						},
					})

				mockClient.EXPECT().
					GetBlockByHashOrNumber("100").
					Return(&domain.BlockResponse{
						Number: 100,
						Timestamp: domain.Timestamp{
							From: "1673222400", // 8 days later
							To:   "1673222401",
						},
					})
			},
			expectOk:       false,
			expectError:    true,
			expectedParams: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			ok, errRpc := commonService.ValidateBlockRangeAndAddTimestampToParams(tc.params, tc.fromBlock, tc.toBlock, tc.address)

			assert.Equal(t, tc.expectOk, ok)
			if tc.expectError {
				assert.NotNil(t, errRpc)
			} else {
				assert.Nil(t, errRpc)
				assert.Equal(t, tc.expectedParams, tc.params)
			}
		})
	}
}

func TestCommonGetLogs(t *testing.T) {
	ctrl, mockClient, _, commonService := setupCommonTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		logParams      domain.LogParams
		mockSetup      func()
		expectedResult []domain.Log
		expectError    bool
	}{
		{
			name: "Success with block hash",
			logParams: domain.LogParams{
				BlockHash: "0x123abc",
				Topics:    []string{"0xtopic1", "0xtopic2"},
			},
			mockSetup: func() {
				mockClient.EXPECT().
					GetBlockByHashOrNumber("0x123abc").
					Return(&domain.BlockResponse{
						Timestamp: domain.Timestamp{
							From: "1672531200",
							To:   "1672531201",
						},
					})

				mockClient.EXPECT().
					GetContractResultsLogsWithRetry(map[string]interface{}{
						"timestamp": "gte:1672531200&timestamp=lte:1672531201",
						"topic0":    "0xtopic1",
						"topic1":    "0xtopic2",
					}).
					Return([]domain.LogEntry{
						{
							Address:          "0xaddress1",
							BlockHash:        "0xblockhash1",
							BlockNumber:      ptr(int64(1)),
							Data:             "0xdata1",
							TransactionHash:  "0xtxhash1",
							TransactionIndex: ptr(0),
							Index:            ptr(0),
							Topics:           []string{"0xtopic1", "0xtopic2"},
						},
					}, nil)
			},
			expectedResult: []domain.Log{
				{
					Address:          "0xaddress1",
					BlockHash:        "0xblockhash1",
					BlockNumber:      "0x1",
					Data:             "0xdata1",
					LogIndex:         "0x0",
					Removed:          false,
					Topics:           []string{"0xtopic1", "0xtopic2"},
					TransactionHash:  "0xtxhash1",
					TransactionIndex: "0x0",
				},
			},
			expectError: false,
		},
		{
			name: "Success with block range",
			logParams: domain.LogParams{
				FromBlock: "0x1",
				ToBlock:   "0x2",
				Address:   []string{"0xaddress1"},
			},
			mockSetup: func() {
				// Mock GetLatestBlock for block range validation
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{"number": float64(100)}, nil)

				// Mock getting from block
				mockClient.EXPECT().
					GetBlockByHashOrNumber("1").
					Return(&domain.BlockResponse{
						Number: 1,
						Timestamp: domain.Timestamp{
							From: "1672531200",
							To:   "1672531201",
						},
					})

				// Mock getting to block
				mockClient.EXPECT().
					GetBlockByHashOrNumber("2").
					Return(&domain.BlockResponse{
						Number: 2,
						Timestamp: domain.Timestamp{
							From: "1672531201",
							To:   "1672531202",
						},
					})

				// Mock getting logs
				mockClient.EXPECT().
					GetContractResultsLogsByAddress("0xaddress1", map[string]interface{}{
						"timestamp": "gte:1672531200&timestamp=lte:1672531202",
					}).
					Return([]domain.LogEntry{
						{
							Address:          "0xaddress1",
							BlockHash:        "0xblockhash1",
							BlockNumber:      ptr(int64(1)),
							Data:             "0xdata1",
							TransactionHash:  "0xtxhash1",
							TransactionIndex: ptr(0),
							Index:            ptr(0),
							Topics:           []string{},
						},
					}, nil)
			},
			expectedResult: []domain.Log{
				{
					Address:          "0xaddress1",
					BlockHash:        "0xblockhash1",
					BlockNumber:      "0x1",
					Data:             "0xdata1",
					LogIndex:         "0x0",
					Removed:          false,
					Topics:           []string{},
					TransactionHash:  "0xtxhash1",
					TransactionIndex: "0x0",
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, errRpc := commonService.GetLogs(tc.logParams)

			if tc.expectError {
				assert.NotNil(t, errRpc)
			} else {
				assert.Nil(t, errRpc)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestCommonGetBlockNumber(t *testing.T) {
	ctrl, mockClient, _, commonService := setupCommonTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		mockSetup      func()
		expectedResult interface{}
		expectError    bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": float64(42),
					}, nil)
			},
			expectedResult: "0x2a", // 42 in hex
			expectError:    false,
		},
		{
			name: "Error getting latest block",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(nil, fmt.Errorf("failed to fetch block"))
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name: "Invalid block number type",
			mockSetup: func() {
				mockClient.EXPECT().
					GetLatestBlock().
					Return(map[string]interface{}{
						"number": "not a number",
					}, nil)
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, errRpc := commonService.GetBlockNumber()

			if tc.expectError {
				assert.NotNil(t, errRpc)
				assert.Equal(t, tc.expectedResult, result)
			} else {
				assert.Nil(t, errRpc)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
