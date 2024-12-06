package limiter

import (
	"sync"
	"time"
)

type TierConfig struct {
	RequestsPerMinute int
	HbarLimit         int
}

type TieredLimiter struct {
	tierConfigs           map[string]*TierConfig
	operatorHbarRemaining int
	mu                    sync.Mutex
	userRequestCounters   map[string]int
	userHbarCounters      map[string]int
	userLastReset         map[string]time.Time
}

func NewTieredLimiter(cfg map[string]interface{}, operatorHbarBudget int) *TieredLimiter {
	tl := &TieredLimiter{
		tierConfigs:           make(map[string]*TierConfig),
		operatorHbarRemaining: operatorHbarBudget,
		userRequestCounters:   make(map[string]int),
		userHbarCounters:      make(map[string]int),
		userLastReset:         make(map[string]time.Time),
	}

	for tierName, val := range cfg {
		if m, ok := val.(map[interface{}]interface{}); ok {
			tl.tierConfigs[tierName] = &TierConfig{
				RequestsPerMinute: m["requestsPerMinute"].(int),
				HbarLimit:         m["hbarLimit"].(int),
			}
		}
	}
	return tl
}

func (t *TieredLimiter) CheckLimits(apiKey string, tier string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	tc, exists := t.tierConfigs[tier]
	if !exists {
		return false
	}

	now := time.Now()
	lastReset, ok := t.userLastReset[apiKey]
	if !ok || now.Sub(lastReset) > time.Minute {
		t.userRequestCounters[apiKey] = 0
		t.userHbarCounters[apiKey] = 0
		t.userLastReset[apiKey] = now
	}

	if t.userRequestCounters[apiKey] >= tc.RequestsPerMinute {
		return false
	}

	t.userRequestCounters[apiKey]++
	return true
}

func (t *TieredLimiter) DeductHbarUsage(apiKey, tier string, amount int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	tc, exists := t.tierConfigs[tier]
	if !exists {
		return false
	}

	if t.userHbarCounters[apiKey]+amount > tc.HbarLimit {
		return false
	}
	if t.operatorHbarRemaining < amount {
		return false
	}

	t.userHbarCounters[apiKey] += amount
	t.operatorHbarRemaining -= amount
	return true
}
