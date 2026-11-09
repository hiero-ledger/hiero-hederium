package limiter_test

import (
	"testing"

	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/limiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The stores read entries in the map[interface{}]interface{} shape of the
// YAML decoder. Viper hands them map[string]interface{} instead, so nothing
// loads from real config today (gap report E-4, fixed separately). These
// tests cover the per-field checks: an entry with a field of the wrong type
// is skipped instead of panicking at start-up.

func TestNewAPIKeyStore_SkipsMalformedEntries(t *testing.T) {
	var store *limiter.APIKeyStore
	require.NotPanics(t, func() {
		store = limiter.NewAPIKeyStore([]interface{}{
			map[interface{}]interface{}{"key": "valid-key", "tier": "free"},
			map[interface{}]interface{}{"key": 12345, "tier": "free"},
			map[interface{}]interface{}{"key": "no-tier"},
			map[interface{}]interface{}{"key": "numeric-tier", "tier": 1},
		})
	})

	tier, ok := store.GetTierForKey("valid-key")
	assert.True(t, ok)
	assert.Equal(t, "free", tier)

	for _, key := range []string{"12345", "no-tier", "numeric-tier"} {
		_, ok := store.GetTierForKey(key)
		assert.False(t, ok, key)
	}
}

func TestNewTieredLimiter_SkipsMalformedTiers(t *testing.T) {
	var tl *limiter.TieredLimiter
	require.NotPanics(t, func() {
		tl = limiter.NewTieredLimiter(map[string]interface{}{
			"free":    map[interface{}]interface{}{"requestsPerMinute": 1, "hbarLimit": 10},
			"quoted":  map[interface{}]interface{}{"requestsPerMinute": "60", "hbarLimit": 10},
			"partial": map[interface{}]interface{}{"requestsPerMinute": 60},
		}, 100)
	})

	assert.True(t, tl.CheckLimits("k1", "free"), "first request within the free tier")
	assert.False(t, tl.CheckLimits("k1", "free"), "second request exceeds 1 per minute")
	assert.False(t, tl.CheckLimits("k2", "quoted"), "a tier with a malformed field is not loaded")
	assert.False(t, tl.CheckLimits("k3", "partial"), "a tier with a missing field is not loaded")
}
