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

func setupFilterTest(t *testing.T) (*gomock.Controller, *mocks.MockMirrorClient, *mocks.MockCacheService, *mocks.MockCommonService, service.FilterService) {
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewDevelopment()
	mockClient := mocks.NewMockMirrorClient(ctrl)
	mockCache := mocks.NewMockCacheService(ctrl)
	mockCommon := mocks.NewMockCommonService(ctrl)
	filterService := service.NewFilterService(mockClient, mockCache, logger, mockCommon)

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
			expectError:    true,
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
