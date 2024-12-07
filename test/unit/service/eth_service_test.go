package service_test

import (
	"testing"

	"github.com/georgi-l95/Hederium/internal/service"
	"github.com/georgi-l95/Hederium/test/unit/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestGetBlockNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Create mock client from the interface
	mockClient := mocks.NewMockMirrorNodeClient(ctrl)
	mockClient.EXPECT().
		GetLatestBlock().
		Return(map[string]interface{}{"number": float64(42)}, nil)

	s := service.NewEthService(
		nil,        // hClient not needed for this test
		mockClient, // pass the mock as the interface
		logger,
		nil, // tieredLimiter not needed for this test
	)

	result, errMap := s.GetBlockNumber()
	assert.Nil(t, errMap)
	// 42 in hex is "0x2a"
	assert.Equal(t, "0x2a", result)
}
