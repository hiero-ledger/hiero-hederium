package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNetService_Listening(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	chainId := "testnet-123"
	netService := NewNetService(logger, chainId)

	// Test
	result := netService.Listening()

	// Assert
	assert.False(t, result, "Listening should always return false for Hedera network")
}

func TestNetService_Version(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	expectedChainId := "testnet-123"
	netService := NewNetService(logger, expectedChainId)

	// Test
	result := netService.Version()

	// Assert
	assert.Equal(t, expectedChainId, result, "Version should return the chain ID")
}
