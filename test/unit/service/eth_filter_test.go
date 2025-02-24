package service

import (
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
