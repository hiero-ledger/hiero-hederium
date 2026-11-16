package startup_test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/config"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/startup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout returns what fn printed to standard output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	defer func() { os.Stdout = original }()

	printed := make(chan string)
	go func() {
		data, _ := io.ReadAll(reader)
		printed <- string(data)
	}()

	fn()
	require.NoError(t, writer.Close())
	return <-printed
}

func TestLogStartup_PrintsSettingsAndNoSecrets(t *testing.T) {
	const operatorKey = "fake-operator-key-for-tests"
	const apiKey = "fake-api-key-for-tests"
	cfg := &config.Config{
		Application: config.Application{Version: "1.2.3"},
		Server:      config.Server{Port: 7546},
		Hedera: config.Hedera{
			Network:     "testnet",
			ChainID:     "0x128",
			OperatorID:  "0.0.1001",
			OperatorKey: config.Secret(operatorKey),
		},
		MirrorNode: config.MirrorNode{BaseURL: "https://mirror.example", Timeout: 1500 * time.Millisecond},
		Features:   config.Features{EnableBatchRequests: true},
		Cache:      config.Cache{DefaultExpiration: time.Hour, CleanupInterval: 30 * time.Minute},
		APIKeys: []interface{}{
			map[string]interface{}{"key": apiKey, "tier": "free"},
			apiKey,
		},
	}

	out := captureStdout(t, func() { startup.LogStartup(cfg) })

	assert.Contains(t, out, "Starting Hederium v1.2.3")
	assert.Contains(t, out, "Network: testnet")
	assert.Contains(t, out, "Chain ID: 0x128")
	assert.Contains(t, out, "Port: 7546")
	assert.Contains(t, out, "Enable Batch Requests: true")
	assert.Contains(t, out, "Base URL: https://mirror.example")
	assert.Contains(t, out, "Timeout: 1.5s")
	assert.Contains(t, out, "Default Expiration: 1h0m0s")

	assert.NotContains(t, out, operatorKey)
	assert.NotContains(t, out, apiKey)
}
