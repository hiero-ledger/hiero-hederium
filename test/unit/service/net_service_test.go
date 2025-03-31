package service_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/LimeChain/Hederium/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNetService_Listening(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	chainId := "testnet-123"
	netService := service.NewNetService(logger, chainId)

	// Test
	result := netService.Listening()

	// Assert
	assert.Equal(t, "false", result, "Listening should always return false for Hedera network")
}

func TestNetService_Version(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	expectedChainId := "0x12a"
	netService := service.NewNetService(logger, expectedChainId)

	// Test
	result := netService.Version()
	digit, _ := strconv.ParseInt(result, 10, 64)
	result = fmt.Sprintf("0x%x", digit)

	// Assert
	assert.Equal(t, expectedChainId, result, "Version should return the chain ID")
}
