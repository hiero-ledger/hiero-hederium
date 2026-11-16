package config_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeKey stands in for the operator key. It is not a key.
const fakeKey = "fake-operator-key-for-tests"

// mandatoryEnv is the smallest environment Load accepts, in relay names.
var mandatoryEnv = map[string]string{
	"HEDERA_NETWORK":    "testnet",
	"CHAIN_ID":          "296",
	"MIRROR_NODE_URL":   "https://mirror.example",
	"OPERATOR_ID_MAIN":  "0.0.1001",
	"OPERATOR_KEY_MAIN": fakeKey,
}

// fullFile sets every key, with numbers quoted the way the Helm chart renders them.
const fullFile = `
application:
  version: "9.9.9"
server:
  port: 8080
hedera:
  network: "previewnet"
  operatorId: "0.0.2002"
  operatorKey: "` + fakeKey + `"
  chainId: "0x129"
  hbarBudget: "1000"
mirrorNode:
  baseUrl: "http://127.0.0.1:5551"
  timeoutSeconds: "7"
limiter:
  free:
    requestsPerMinute: 100
    hbarLimit: 10
logging:
  level: "warn"
apiKeys:
  - key: "fake-api-key"
    tier: "free"
features:
  enforceApiKey: true
  enableBatchRequests: false
cache:
  defaultExpiration: "2h"
  cleanupInterval: "15m"
`

// isolate clears every variable Load reads, so the shell that runs the tests
// cannot change their outcome. An empty variable counts as not set.
func isolate(t *testing.T) {
	t.Helper()
	for _, name := range config.EnvNames() {
		t.Setenv(name, "")
	}
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for name, value := range env {
		t.Setenv(name, value)
	}
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600))
	return dir
}

// noFile is a directory without a config.yaml.
func noFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "missing")
}

func TestLoad_EnvironmentOnlyUsesRelayNamesAndDefaults(t *testing.T) {
	isolate(t)
	setEnv(t, mandatoryEnv)

	cfg, err := config.LoadFrom(noFile(t))
	require.NoError(t, err)

	assert.Equal(t, "testnet", cfg.Hedera.Network)
	assert.Equal(t, "0x128", cfg.Hedera.ChainID)
	assert.Equal(t, "https://mirror.example", cfg.MirrorNode.BaseURL)
	assert.Equal(t, "0.0.1001", cfg.Hedera.OperatorID)
	assert.Equal(t, fakeKey, cfg.Hedera.OperatorKey.Reveal())

	assert.Equal(t, "0.1.0", cfg.Application.Version)
	assert.Equal(t, 7546, cfg.Server.Port)
	assert.Equal(t, 0, cfg.Hedera.HbarBudget)
	assert.Equal(t, 10*time.Second, cfg.MirrorNode.Timeout)
	assert.Equal(t, "debug", cfg.Logging.Level, "the relay's default level, trace, runs as debug")
	assert.False(t, cfg.Features.EnforceAPIKey)
	assert.True(t, cfg.Features.EnableBatchRequests)
	assert.Equal(t, time.Hour, cfg.Cache.DefaultExpiration)
	assert.Equal(t, 30*time.Minute, cfg.Cache.CleanupInterval)
	assert.Nil(t, cfg.APIKeys)
	assert.Empty(t, cfg.Limiter)
}

func TestLoad_FileOnly(t *testing.T) {
	isolate(t)

	cfg, err := config.LoadFrom(writeFile(t, fullFile))
	require.NoError(t, err)

	assert.Equal(t, "9.9.9", cfg.Application.Version)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "previewnet", cfg.Hedera.Network)
	assert.Equal(t, "0.0.2002", cfg.Hedera.OperatorID)
	assert.Equal(t, fakeKey, cfg.Hedera.OperatorKey.Reveal())
	assert.Equal(t, "0x129", cfg.Hedera.ChainID)
	assert.Equal(t, 1000, cfg.Hedera.HbarBudget)
	assert.Equal(t, "http://127.0.0.1:5551", cfg.MirrorNode.BaseURL)
	assert.Equal(t, 7*time.Second, cfg.MirrorNode.Timeout)
	assert.Equal(t, "warn", cfg.Logging.Level)
	assert.True(t, cfg.Features.EnforceAPIKey)
	assert.False(t, cfg.Features.EnableBatchRequests)
	assert.Equal(t, 2*time.Hour, cfg.Cache.DefaultExpiration)
	assert.Equal(t, 15*time.Minute, cfg.Cache.CleanupInterval)

	keys, ok := cfg.APIKeys.([]interface{})
	require.True(t, ok, "apiKeys is passed through as the list from the file")
	assert.Len(t, keys, 1)
	assert.Contains(t, cfg.Limiter, "free")
}

func TestLoad_StrongestSourceWins(t *testing.T) {
	dir := writeFile(t, fullFile)
	load := func(t *testing.T) *config.Config {
		t.Helper()
		cfg, err := config.LoadFrom(dir)
		require.NoError(t, err)
		return cfg
	}

	t.Run("file when no variable is set", func(t *testing.T) {
		isolate(t)
		assert.Equal(t, "http://127.0.0.1:5551", load(t).MirrorNode.BaseURL)
	})

	t.Run("dotted Hederium variable beats the file", func(t *testing.T) {
		isolate(t)
		t.Setenv("MIRRORNODE.BASEURL", "http://dotted.example")
		assert.Equal(t, "http://dotted.example", load(t).MirrorNode.BaseURL)
	})

	t.Run("underscore Hederium variable beats the dotted one", func(t *testing.T) {
		isolate(t)
		t.Setenv("MIRRORNODE.BASEURL", "http://dotted.example")
		t.Setenv("MIRRORNODE_BASEURL", "http://own.example")
		assert.Equal(t, "http://own.example", load(t).MirrorNode.BaseURL)
	})

	t.Run("relay variable beats everything", func(t *testing.T) {
		isolate(t)
		t.Setenv("MIRRORNODE.BASEURL", "http://dotted.example")
		t.Setenv("MIRRORNODE_BASEURL", "http://own.example")
		t.Setenv("MIRROR_NODE_URL", "http://relay.example")
		assert.Equal(t, "http://relay.example", load(t).MirrorNode.BaseURL)
	})

	t.Run("an empty variable counts as not set", func(t *testing.T) {
		isolate(t)
		t.Setenv("MIRROR_NODE_URL", "   ")
		assert.Equal(t, "http://127.0.0.1:5551", load(t).MirrorNode.BaseURL)
	})
}

func TestLoad_EverySettingHasAnEnvironmentVariable(t *testing.T) {
	isolate(t)
	setEnv(t, mandatoryEnv)
	setEnv(t, map[string]string{
		"APPLICATION_VERSION":    "1.2.3",
		"SERVER_PORT":            "9000",
		"HEDERA_HBARBUDGET":      "5",
		"LOG_LEVEL":              "error",
		"FEATURES_ENFORCEAPIKEY": "true",
		"BATCH_REQUESTS_ENABLED": "false",
		"CACHE_CLEANUPINTERVAL":  "5m",
	})

	cfg, err := config.LoadFrom(noFile(t))
	require.NoError(t, err)

	assert.Equal(t, "1.2.3", cfg.Application.Version)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, 5, cfg.Hedera.HbarBudget)
	assert.Equal(t, "error", cfg.Logging.Level)
	assert.True(t, cfg.Features.EnforceAPIKey)
	assert.False(t, cfg.Features.EnableBatchRequests)
	assert.Equal(t, 5*time.Minute, cfg.Cache.CleanupInterval)
}

func TestLoad_DurationUnitFollowsTheName(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		timeout   time.Duration
		cacheTime time.Duration
	}{
		{
			name:      "relay names are milliseconds",
			env:       map[string]string{"MIRROR_NODE_TIMEOUT": "500", "CACHE_TTL": "60000"},
			timeout:   500 * time.Millisecond,
			cacheTime: time.Minute,
		},
		{
			name:      "Hederium names keep the units of the file",
			env:       map[string]string{"MIRRORNODE_TIMEOUTSECONDS": "3", "CACHE_DEFAULTEXPIRATION": "90s"},
			timeout:   3 * time.Second,
			cacheTime: 90 * time.Second,
		},
		{
			name:      "a cache time of zero is allowed: entries never expire",
			env:       map[string]string{"CACHE_TTL": "0"},
			timeout:   10 * time.Second,
			cacheTime: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolate(t)
			setEnv(t, mandatoryEnv)
			setEnv(t, tt.env)

			cfg, err := config.LoadFrom(noFile(t))
			require.NoError(t, err)
			assert.Equal(t, tt.timeout, cfg.MirrorNode.Timeout)
			assert.Equal(t, tt.cacheTime, cfg.Cache.DefaultExpiration)
		})
	}
}

func TestLoad_ChainIDIsNormalisedToHex(t *testing.T) {
	for raw, want := range map[string]string{
		"298":   "0x12a",
		"0x12a": "0x12a",
		"0x12A": "0x12a",
		"0X128": "0x128",
		"0x127": "0x127",
		"1":     "0x1",
	} {
		t.Run(raw, func(t *testing.T) {
			isolate(t)
			setEnv(t, mandatoryEnv)
			t.Setenv("CHAIN_ID", raw)

			cfg, err := config.LoadFrom(noFile(t))
			require.NoError(t, err)
			assert.Equal(t, want, cfg.Hedera.ChainID)
		})
	}
}

func TestLoad_LogLevelAcceptsRelayNames(t *testing.T) {
	for raw, want := range map[string]string{
		"trace":  "debug",
		"debug":  "debug",
		"info":   "info",
		"WARN":   "warn",
		"error":  "error",
		"fatal":  "fatal",
		"silent": "fatal",
	} {
		t.Run(raw, func(t *testing.T) {
			isolate(t)
			setEnv(t, mandatoryEnv)
			t.Setenv("LOG_LEVEL", raw)

			cfg, err := config.LoadFrom(noFile(t))
			require.NoError(t, err)
			assert.Equal(t, want, cfg.Logging.Level)
		})
	}
}

func TestLoad_BooleansIgnoreCase(t *testing.T) {
	isolate(t)
	setEnv(t, mandatoryEnv)
	t.Setenv("BATCH_REQUESTS_ENABLED", "False")
	t.Setenv("FEATURES_ENFORCEAPIKEY", "TRUE")

	cfg, err := config.LoadFrom(noFile(t))
	require.NoError(t, err)
	assert.False(t, cfg.Features.EnableBatchRequests)
	assert.True(t, cfg.Features.EnforceAPIKey)
}

func TestLoad_MissingMandatorySettingIsNamed(t *testing.T) {
	keys := map[string]string{
		"HEDERA_NETWORK":    "hedera.network",
		"CHAIN_ID":          "hedera.chainId",
		"MIRROR_NODE_URL":   "mirrorNode.baseUrl",
		"OPERATOR_ID_MAIN":  "hedera.operatorId",
		"OPERATOR_KEY_MAIN": "hedera.operatorKey",
	}
	for missing, key := range keys {
		t.Run(missing, func(t *testing.T) {
			isolate(t)
			setEnv(t, mandatoryEnv)
			t.Setenv(missing, "")

			cfg, err := config.LoadFrom(noFile(t))
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), missing+" is mandatory")
			assert.Contains(t, err.Error(), key+" in config.yaml")
			for other := range keys {
				if other != missing {
					assert.NotContains(t, err.Error(), other+" is mandatory")
				}
			}
		})
	}

	t.Run("all of them are reported at once", func(t *testing.T) {
		isolate(t)

		_, err := config.LoadFrom(noFile(t))
		require.Error(t, err)
		for name := range keys {
			assert.Contains(t, err.Error(), name+" is mandatory")
		}
	})
}

func TestLoad_BadValueIsRefusedAndNamed(t *testing.T) {
	tests := []struct {
		variable string
		values   []string
		message  string
	}{
		{"SERVER_PORT", []string{"abc", "0", "70000", "7546.5", "-1"}, "SERVER_PORT must be a whole number from 1 to 65535"},
		{"HEDERA_HBARBUDGET", []string{"-1", "many"}, "HEDERA_HBARBUDGET must be a whole number, 0 or more"},
		{"CHAIN_ID", []string{"abc", "0x", "0", "0x0", "-1", "12a", "1.5", "0xzz"}, "CHAIN_ID must be a chain ID above zero"},
		{"MIRROR_NODE_URL", []string{"mirror.example", "ftp://mirror.example", "http://", "://x"}, "MIRROR_NODE_URL must be an http or https URL"},
		{"MIRROR_NODE_TIMEOUT", []string{"ten", "-1", "1.5", "10s", "99999999999999999999"}, "MIRROR_NODE_TIMEOUT must be a whole number of milliseconds"},
		{"MIRRORNODE_TIMEOUTSECONDS", []string{"ten", "-1", "10s", "9223372036854775807"}, "MIRRORNODE_TIMEOUTSECONDS must be a whole number of seconds"},
		{"MIRROR_NODE_TIMEOUT", []string{"0"}, "MIRROR_NODE_TIMEOUT must be above zero"},
		{"MIRRORNODE_TIMEOUTSECONDS", []string{"0"}, "MIRRORNODE_TIMEOUTSECONDS must be above zero"},
		{"CACHE_TTL", []string{"1h", "-5"}, "CACHE_TTL must be a whole number of milliseconds"},
		{"CACHE_DEFAULTEXPIRATION", []string{"100", "-1h", "soon"}, "CACHE_DEFAULTEXPIRATION must be a duration such as 90s or 1h"},
		{"CACHE_CLEANUPINTERVAL", []string{"30"}, "CACHE_CLEANUPINTERVAL must be a duration such as 90s or 1h"},
		{"BATCH_REQUESTS_ENABLED", []string{"yes", "1", "on"}, "BATCH_REQUESTS_ENABLED must be true or false"},
		{"FEATURES_ENFORCEAPIKEY", []string{"no"}, "FEATURES_ENFORCEAPIKEY must be true or false"},
		{"LOG_LEVEL", []string{"loud", "verbose"}, "LOG_LEVEL must be one of trace, debug, info, warn, error, fatal, silent"},
	}
	for _, tt := range tests {
		for _, value := range tt.values {
			t.Run(tt.variable+"="+value, func(t *testing.T) {
				isolate(t)
				setEnv(t, mandatoryEnv)
				t.Setenv(tt.variable, value)

				cfg, err := config.LoadFrom(noFile(t))
				require.Error(t, err)
				assert.Nil(t, cfg)
				assert.Contains(t, err.Error(), tt.message)
				assert.NotContains(t, err.Error(), "is mandatory", "a bad value is not reported as missing too")
			})
		}
	}
}

func TestLoad_BadValueInFileNamesTheKey(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		message string
	}{
		{"wrong unit", "mirrorNode:\n  timeoutSeconds: 10s\n", "mirrorNode.timeoutSeconds in config.yaml must be a whole number of seconds"},
		{"duration without a unit", "cache:\n  defaultExpiration: 100\n", "cache.defaultExpiration in config.yaml must be a duration such as 90s or 1h"},
		{"list where one value is expected", "server:\n  port: [7546, 7547]\n", "server.port in config.yaml must be a single value"},
		{"map where one value is expected", "hedera:\n  network:\n    name: testnet\n", "hedera.network in config.yaml must be a single value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolate(t)
			setEnv(t, mandatoryEnv)
			if tt.name == "map where one value is expected" {
				t.Setenv("HEDERA_NETWORK", "")
			}

			_, err := config.LoadFrom(writeFile(t, tt.file))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.message)
			assert.NotContains(t, err.Error(), "is mandatory")
		})
	}
}

func TestLoad_KeyWithoutAValueInFileCountsAsNotSet(t *testing.T) {
	isolate(t)
	setEnv(t, mandatoryEnv)
	t.Setenv("OPERATOR_KEY_MAIN", "")

	_, err := config.LoadFrom(writeFile(t, "hedera:\n  operatorKey:\nserver:\n  port:\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OPERATOR_KEY_MAIN is mandatory")
	assert.NotContains(t, err.Error(), "server.port")
}

func TestLoad_ReadsConfigsDirectoryOfTheWorkingDirectory(t *testing.T) {
	isolate(t)
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "configs"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "configs", "config.yaml"), []byte(fullFile), 0o600))

	original, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	defer func() { require.NoError(t, os.Chdir(original)) }()

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoad_ReportsEveryProblemAtOnce(t *testing.T) {
	isolate(t)
	setEnv(t, mandatoryEnv)
	t.Setenv("SERVER_PORT", "abc")
	t.Setenv("LOG_LEVEL", "loud")
	t.Setenv("OPERATOR_ID_MAIN", "")

	_, err := config.LoadFrom(noFile(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_PORT must be")
	assert.Contains(t, err.Error(), "LOG_LEVEL must be")
	assert.Contains(t, err.Error(), "OPERATOR_ID_MAIN is mandatory")
}

func TestLoad_UnreadableFileIsAnError(t *testing.T) {
	isolate(t)
	setEnv(t, mandatoryEnv)

	_, err := config.LoadFrom(writeFile(t, "server:\n  port: [unclosed\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error reading config")
}

func TestLoad_ErrorsNeverContainAValue(t *testing.T) {
	// The operator key pasted into each setting in turn: the error names the
	// setting and must not repeat what was in it.
	for _, name := range config.EnvNames() {
		t.Run(name, func(t *testing.T) {
			isolate(t)
			t.Setenv(name, fakeKey)

			_, err := config.LoadFrom(noFile(t))
			require.Error(t, err, "the other mandatory settings are missing")
			assert.NotContains(t, err.Error(), fakeKey)
		})
	}

	t.Run("from the file", func(t *testing.T) {
		isolate(t)
		file := "server:\n  port: " + fakeKey + "\nhedera:\n  chainId: " + fakeKey + "\n  operatorKey: " + fakeKey + "\n"

		_, err := config.LoadFrom(writeFile(t, file))
		require.Error(t, err)
		assert.NotContains(t, err.Error(), fakeKey)
	})
}

func TestSecret_IsNotPrintedOrEncoded(t *testing.T) {
	isolate(t)
	setEnv(t, mandatoryEnv)
	cfg, err := config.LoadFrom(noFile(t))
	require.NoError(t, err)
	require.Equal(t, fakeKey, cfg.Hedera.OperatorKey.Reveal())

	encoded, err := json.Marshal(cfg)
	require.NoError(t, err)
	printed := []string{
		fmt.Sprint(cfg.Hedera.OperatorKey),
		fmt.Sprintf("%s", cfg.Hedera.OperatorKey),
		fmt.Sprintf("%q", cfg.Hedera.OperatorKey),
		fmt.Sprintf("%v", cfg),
		fmt.Sprintf("%+v", *cfg),
		fmt.Sprintf("%#v", *cfg),
		fmt.Sprintf("%v", cfg.Hedera),
		fmt.Sprintf("%v", &cfg.Hedera.OperatorKey),
		string(encoded),
	}
	for i, text := range printed {
		assert.NotContains(t, text, fakeKey, "form %d", i)
	}
	assert.Contains(t, string(encoded), "[REDACTED]")
	assert.Contains(t, fmt.Sprintf("%+v", *cfg), "[REDACTED]")
}

func TestDocumentation_NamesEverySetting(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "docs", "configuration.md"))
	require.NoError(t, err)

	for _, name := range config.EnvNames() {
		if strings.Contains(name, ".") {
			continue // the dotted forms are documented by their rule, not listed
		}
		assert.True(t, strings.Contains(string(doc), "`"+name+"`"), "docs/configuration.md does not mention %s", name)
	}
}
