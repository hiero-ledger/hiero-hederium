package service

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

func setupFilterTest(t *testing.T) (*gomock.Controller, *mocks.MockMirrorClient, *mocks.MockCacheService, *mocks.MockCommonService, service.FilterServicer) {
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	mockCache := mocks.NewMockCacheService(ctrl)
	mockCommon := mocks.NewMockCommonService(ctrl)
	filterService := service.NewFilterService(mockClient, mockCache, logger, mockCommon, true)

	return ctrl, mockClient, mockCache, mockCommon, filterService
}

func TestNewFilter(t *testing.T) {
	ctrl, _, mockCache, mockCommon, service := setupFilterTest(t)
	defer ctrl.Finish()

	t.Run("Success_with_valid_block_range", func(t *testing.T) {
		// Mock ValidateBlockRange
		mockCommon.EXPECT().ValidateBlockRange("latest", "latest").Return(nil)

		// Mock GetBlockNumberByNumberOrTag for "latest" in NewFilter
		mockCommon.EXPECT().GetBlockNumberByNumberOrTag("latest").Return(int64(100), nil)

		// Mock cache Set
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		filter, err := service.NewFilter("latest", "latest", []string{"0xaddress"}, []string{"0xtopic1"})

		assert.Nil(t, err)
		assert.NotNil(t, filter)
	})

	t.Run("Error_with_invalid_block_range", func(t *testing.T) {
		// Mock ValidateBlockRange to return error
		mockCommon.EXPECT().ValidateBlockRange("0x2", "0x1").Return(domain.NewInvalidBlockRangeError())

		filter, err := service.NewFilter("0x2", "0x1", []string{"0xaddress"}, []string{"0xtopic1"})

		assert.NotNil(t, err)
		assert.Nil(t, filter)
		assert.Equal(t, "Invalid block range", err.Message)
	})
}

func TestNewBlockFilter(t *testing.T) {
	ctrl, _, mockCache, mockCommon, filterService := setupFilterTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		mockSetup      func()
		expectError    bool
		expectedResult bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockCommon.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(100), nil)

				mockCache.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError:    false,
			expectedResult: true,
		},
		{
			name: "Error getting block number",
			mockSetup: func() {
				mockCommon.EXPECT().
					GetBlockNumberByNumberOrTag("latest").
					Return(int64(0), domain.NewRPCError(domain.ServerError, "failed to get block"))
			},
			expectError:    true,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, errRpc := filterService.NewBlockFilter()

			if tc.expectError {
				assert.NotNil(t, errRpc)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, errRpc)
				assert.NotNil(t, result)
				assert.Equal(t, 34, len(*result)) // "0x" + 16 bytes hex
			}
		})
	}
}

func TestUninstallFilter(t *testing.T) {
	ctrl, _, mockCache, _, filterService := setupFilterTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		filterID       string
		mockSetup      func()
		expectError    bool
		expectedResult bool
	}{
		{
			name:     "Success",
			filterID: "0x123abc",
			mockSetup: func() {
				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc"), gomock.Any()).
					Return(nil)

				mockCache.EXPECT().
					Delete(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc")).
					Return(nil)
			},
			expectError:    false,
			expectedResult: true,
		},
		{
			name:     "Filter not found",
			filterID: "0xnonexistent",
			mockSetup: func() {
				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0xnonexistent"), gomock.Any()).
					Return(fmt.Errorf("not found"))
			},
			expectError:    false,
			expectedResult: false,
		},
		{
			name:     "Error deleting filter",
			filterID: "0x123abc",
			mockSetup: func() {
				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc"), gomock.Any()).
					Return(nil)

				mockCache.EXPECT().
					Delete(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc")).
					Return(fmt.Errorf("delete error"))
			},
			expectError:    true,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, errRpc := filterService.UninstallFilter(tc.filterID)

			if tc.expectError {
				assert.NotNil(t, errRpc)
			} else {
				assert.Nil(t, errRpc)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestNewPendingTransactionFilter(t *testing.T) {
	ctrl, _, _, _, filterService := setupFilterTest(t)
	defer ctrl.Finish()

	result, errRpc := filterService.NewPendingTransactionFilter()
	assert.Nil(t, result)
	assert.NotNil(t, errRpc)
	assert.Equal(t, -32601, errRpc.Code)
}

func TestGetFilterLogs(t *testing.T) {
	ctrl, _, mockCache, mockCommon, filterService := setupFilterTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		filterID       string
		mockSetup      func()
		expectError    bool
		expectedResult []domain.Log
	}{
		{
			name:     "Success",
			filterID: "0x123abc",
			mockSetup: func() {
				filter := domain.Filter{
					ID:        "0x123abc",
					Type:      "log",
					FromBlock: "0x1",
					ToBlock:   "0x2",
					Address:   []string{"0xaddress1"},
					Topics:    []string{"0xtopic1"},
				}

				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc"), gomock.Any()).
					DoAndReturn(func(ctx interface{}, key string, value interface{}) error {
						f := value.(*domain.Filter)
						*f = filter
						return nil
					})

				expectedLogs := []domain.Log{
					{
						Address:          "0xaddress1",
						BlockHash:        "0xblockhash1",
						BlockNumber:      "0x1",
						Data:             "0xdata1",
						LogIndex:         "0x0",
						Topics:           []string{"0xtopic1"},
						TransactionHash:  "0xtxhash1",
						TransactionIndex: "0x0",
					},
				}

				mockCommon.EXPECT().
					GetLogs(gomock.Any()).
					Return(expectedLogs, nil)

				mockCache.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
			expectedResult: []domain.Log{
				{
					Address:          "0xaddress1",
					BlockHash:        "0xblockhash1",
					BlockNumber:      "0x1",
					Data:             "0xdata1",
					LogIndex:         "0x0",
					Topics:           []string{"0xtopic1"},
					TransactionHash:  "0xtxhash1",
					TransactionIndex: "0x0",
				},
			},
		},
		{
			name:     "Filter not found",
			filterID: "0xnonexistent",
			mockSetup: func() {
				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0xnonexistent"), gomock.Any()).
					Return(fmt.Errorf("not found"))
			},
			expectError:    true,
			expectedResult: nil,
		},
		{
			name:     "Invalid filter type",
			filterID: "0x123abc",
			mockSetup: func() {
				filter := domain.Filter{
					ID:   "0x123abc",
					Type: "block", // Not a log filter
				}

				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc"), gomock.Any()).
					DoAndReturn(func(ctx interface{}, key string, value interface{}) error {
						f := value.(*domain.Filter)
						*f = filter
						return nil
					})
			},
			expectError:    true,
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, errRpc := filterService.GetFilterLogs(tc.filterID)

			if tc.expectError {
				assert.NotNil(t, errRpc)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, errRpc)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetFilterChanges(t *testing.T) {
	ctrl, mockClient, mockCache, mockCommon, filterService := setupFilterTest(t)
	defer ctrl.Finish()

	testCases := []struct {
		name           string
		filterID       string
		mockSetup      func()
		expectError    bool
		expectedResult interface{}
	}{
		{
			name:     "Success with log filter",
			filterID: "0x123abc",
			mockSetup: func() {
				filter := domain.Filter{
					ID:        "0x123abc",
					Type:      "log",
					FromBlock: "0x1",
					ToBlock:   "0x2",
					Address:   []string{"0xaddress1"},
					Topics:    []string{"0xtopic1"},
				}

				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc"), gomock.Any()).
					DoAndReturn(func(ctx interface{}, key string, value interface{}) error {
						f := value.(*domain.Filter)
						*f = filter
						return nil
					})

				expectedLogs := []domain.Log{
					{
						Address:          "0xaddress1",
						BlockHash:        "0xblockhash1",
						BlockNumber:      "0x1",
						Data:             "0xdata1",
						LogIndex:         "0x0",
						Topics:           []string{"0xtopic1"},
						TransactionHash:  "0xtxhash1",
						TransactionIndex: "0x0",
					},
				}

				mockCommon.EXPECT().
					GetLogs(gomock.Any()).
					Return(expectedLogs, nil)

				mockCache.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
			expectedResult: []domain.Log{
				{
					Address:          "0xaddress1",
					BlockHash:        "0xblockhash1",
					BlockNumber:      "0x1",
					Data:             "0xdata1",
					LogIndex:         "0x0",
					Topics:           []string{"0xtopic1"},
					TransactionHash:  "0xtxhash1",
					TransactionIndex: "0x0",
				},
			},
		},
		{
			name:     "Success with block filter",
			filterID: "0x123abc",
			mockSetup: func() {
				filter := domain.Filter{
					ID:              "0x123abc",
					Type:            "new_block",
					BlockAtCreation: "0x1",
				}

				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0x123abc"), gomock.Any()).
					DoAndReturn(func(ctx interface{}, key string, value interface{}) error {
						f := value.(*domain.Filter)
						*f = filter
						return nil
					})

				mockClient.EXPECT().
					GetBlocks("0x1").
					Return([]map[string]interface{}{
						{"hash": "0xblockhash1", "number": float64(1)},
						{"hash": "0xblockhash2", "number": float64(2)},
					}, nil)

				mockCache.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError:    false,
			expectedResult: []string{"0xblockhash1", "0xblockhash2"},
		},
		{
			name:     "Filter not found",
			filterID: "0xnonexistent",
			mockSetup: func() {
				mockCache.EXPECT().
					Get(gomock.Any(), fmt.Sprintf("filterId_%s", "0xnonexistent"), gomock.Any()).
					Return(fmt.Errorf("not found"))
			},
			expectError:    true,
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			result, errRpc := filterService.GetFilterChanges(tc.filterID)

			if tc.expectError {
				assert.NotNil(t, errRpc)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, errRpc)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
