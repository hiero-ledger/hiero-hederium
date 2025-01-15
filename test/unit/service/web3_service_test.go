package service_test

import (
	"testing"

	"github.com/LimeChain/Hederium/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWeb3Service_ClientVersion_WithVersion(t *testing.T) {
	// Create a logger for testing (in-memory, no output)
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Given a known version
	version := "1.0.0"
	w := service.NewWeb3Service(logger, version)

	// When calling ClientVersion
	result := w.ClientVersion()

	// Then it should return "relay/1.0.0"
	assert.Equal(t, "relay/1.0.0", result)
}

func TestWeb3Service_ClientVersion_NoVersion(t *testing.T) {
	// Create a logger for testing
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()

	// Given no application version
	version := ""
	w := service.NewWeb3Service(logger, version)

	// When calling ClientVersion
	result := w.ClientVersion()

	// Then it should return "relay/unknown"
	assert.Equal(t, "relay/unknown", result)
}
