package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Event type constants for testing
const (
	testEventTypeBlock = "newHeads"
	testEventTypeLogs  = "logs"
)

// TestPollerService is a simple implementation of PollerService for testing
type TestPollerService struct {
	polls      map[string]*service.Poll
	pollsMutex sync.RWMutex
	isPolling  bool
}

func NewTestPollerService() *TestPollerService {
	return &TestPollerService{
		polls:     make(map[string]*service.Poll),
		isPolling: false,
	}
}

func (p *TestPollerService) Start() {
	p.isPolling = true
}

func (p *TestPollerService) Stop() {
	p.isPolling = false
}

func (p *TestPollerService) AddPoll(tag string, callback service.PollCallback, filters *service.PollFilters) error {
	p.pollsMutex.Lock()
	defer p.pollsMutex.Unlock()

	if poll, exists := p.polls[tag]; exists {
		poll.SubscriberCount++
		return nil
	}

	if callback == nil {
		return errors.New("cannot add poll without callback")
	}

	p.polls[tag] = &service.Poll{
		Tag:             tag,
		Callback:        callback,
		SubscriberCount: 1,
	}

	if !p.isPolling {
		p.Start()
	}

	return nil
}

func (p *TestPollerService) RemoveSubscriptionFromPoll(tag string) {
	p.pollsMutex.Lock()
	defer p.pollsMutex.Unlock()

	if poll, exists := p.polls[tag]; exists {
		poll.SubscriberCount--
		if poll.SubscriberCount <= 0 {
			delete(p.polls, tag)
		}
	}

	if len(p.polls) == 0 {
		p.Stop()
	}
}

func (p *TestPollerService) IsPolling() bool {
	return p.isPolling
}

func (p *TestPollerService) HasPoll(tag string) bool {
	p.pollsMutex.RLock()
	defer p.pollsMutex.RUnlock()

	_, exists := p.polls[tag]
	return exists
}

func (p *TestPollerService) GetPoll(tag string) *service.Poll {
	p.pollsMutex.RLock()
	defer p.pollsMutex.RUnlock()

	return p.polls[tag]
}

// TestCacheService is a simple implementation of CacheService for testing
type TestCacheService struct {
	cache      map[string]interface{}
	cacheMutex sync.RWMutex
}

func NewTestCacheService() *TestCacheService {
	return &TestCacheService{
		cache: make(map[string]interface{}),
	}
}

func (c *TestCacheService) Get(ctx context.Context, key string, value interface{}) error {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	if val, exists := c.cache[key]; exists {
		// This is a simplified implementation for testing
		// In a real implementation, we would unmarshal the value
		switch v := value.(type) {
		case *bool:
			*v = val.(bool)
		default:
			return errors.New("unsupported type")
		}
		return nil
	}

	return errors.New("key not found")
}

func (c *TestCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache[key] = value
	return nil
}

func (c *TestCacheService) Delete(ctx context.Context, key string) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	delete(c.cache, key)
	return nil
}

func setupSubscribeTest(t *testing.T) (*TestPollerService, *TestCacheService, service.SubscribeServicer) {
	pollerService := NewTestPollerService()
	cacheService := NewTestCacheService()

	logger, _ := zap.NewDevelopment()
	subscribeService := service.NewSubscribeService(pollerService, logger, cacheService)

	return pollerService, cacheService, subscribeService
}

func TestNewSubscribeService(t *testing.T) {
	t.Run("creates service with provided dependencies", func(t *testing.T) {
		pollerService := NewTestPollerService()
		cacheService := NewTestCacheService()
		logger, _ := zap.NewDevelopment()

		subscribeService := service.NewSubscribeService(pollerService, logger, cacheService)
		assert.NotNil(t, subscribeService)
	})
}

func TestSubscribeService_Subscribe(t *testing.T) {
	t.Run("Subscribe to logs event", func(t *testing.T) {
		pollerService, _, subscribeService := setupSubscribeTest(t)

		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}

		// Call the method
		callback := func(subscriptionID string, result interface{}) {
			// Callback implementation
		}

		subID, err := subscribeService.Subscribe(testEventTypeLogs, subscribeOptions, callback)

		// Assertions
		assert.NoError(t, err)
		assert.NotEmpty(t, subID)

		// Verify that the poll was added to the poller service
		tagData := struct {
			Event               string   `json:"event"`
			Address             []string `json:"address,omitempty"`
			Topics              []string `json:"topics,omitempty"`
			IncludeTransactions bool     `json:"includeTransactions,omitempty"`
		}{
			Event:   testEventTypeLogs,
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}
		tagJSON, _ := json.Marshal(tagData)
		expectedTag := string(tagJSON)

		assert.True(t, pollerService.HasPoll(expectedTag))
	})

	t.Run("Subscribe to existing poll", func(t *testing.T) {
		pollerService, _, subscribeService := setupSubscribeTest(t)

		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}

		// Create the tag that will be used
		tagData := struct {
			Event               string   `json:"event"`
			Address             []string `json:"address,omitempty"`
			Topics              []string `json:"topics,omitempty"`
			IncludeTransactions bool     `json:"includeTransactions,omitempty"`
		}{
			Event:   testEventTypeLogs,
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}
		tagJSON, _ := json.Marshal(tagData)
		expectedTag := string(tagJSON)

		// Add a poll first
		callback1 := func(subscriptionID string, result interface{}) {}
		subID1, _ := subscribeService.Subscribe(testEventTypeLogs, subscribeOptions, callback1)
		assert.NotEmpty(t, subID1)

		// Now subscribe again with the same options
		callback2 := func(subscriptionID string, result interface{}) {}
		subID2, err := subscribeService.Subscribe(testEventTypeLogs, subscribeOptions, callback2)

		// Assertions
		assert.NoError(t, err)
		assert.NotEmpty(t, subID2)
		assert.NotEqual(t, subID1, subID2)

		// Verify that the poll exists and has 2 subscribers
		assert.True(t, pollerService.HasPoll(expectedTag))
		poll := pollerService.GetPoll(expectedTag)
		assert.Equal(t, 2, poll.SubscriberCount)
	})

	t.Run("Subscribe with invalid event type", func(t *testing.T) {
		_, _, subscribeService := setupSubscribeTest(t)

		// Setup test data
		subscribeOptions := domain.SubscribeOptions{}

		// Call the method
		callback := func(subscriptionID string, result interface{}) {}
		subID, err := subscribeService.Subscribe("invalid_event", subscribeOptions, callback)

		// Assertions
		assert.Error(t, err)
		assert.Empty(t, subID)
	})
}

func TestSubscribeService_Unsubscribe(t *testing.T) {
	t.Run("Unsubscribe from existing subscription", func(t *testing.T) {
		pollerService, _, subscribeService := setupSubscribeTest(t)

		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}

		// Create the tag that will be used
		tagData := struct {
			Event               string   `json:"event"`
			Address             []string `json:"address,omitempty"`
			Topics              []string `json:"topics,omitempty"`
			IncludeTransactions bool     `json:"includeTransactions,omitempty"`
		}{
			Event:   testEventTypeLogs,
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}
		tagJSON, _ := json.Marshal(tagData)
		expectedTag := string(tagJSON)

		// Call Subscribe to create a subscription
		callback := func(subscriptionID string, result interface{}) {}
		subID, _ := subscribeService.Subscribe(testEventTypeLogs, subscribeOptions, callback)

		// Verify the poll exists
		assert.True(t, pollerService.HasPoll(expectedTag))

		// Call Unsubscribe
		success, err := subscribeService.Unsubscribe(subID)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, success)

		// Verify the poll was removed
		assert.False(t, pollerService.HasPoll(expectedTag))
	})

	t.Run("Unsubscribe from non-existent subscription", func(t *testing.T) {
		_, _, subscribeService := setupSubscribeTest(t)

		// Call Unsubscribe with a non-existent ID
		success, err := subscribeService.Unsubscribe("non-existent-id")

		// Assertions
		assert.Error(t, err)
		assert.False(t, success)
	})
}

func TestSubscribeService_HasSubscription(t *testing.T) {
	t.Run("Check for existing subscription", func(t *testing.T) {
		_, _, subscribeService := setupSubscribeTest(t)

		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}

		// Call Subscribe to create a subscription
		callback := func(subscriptionID string, result interface{}) {}
		subID, _ := subscribeService.Subscribe(testEventTypeLogs, subscribeOptions, callback)

		// Check if subscription exists
		exists := subscribeService.HasSubscription(subID)

		// Assertions
		assert.True(t, exists)
	})

	t.Run("Check for non-existent subscription", func(t *testing.T) {
		_, _, subscribeService := setupSubscribeTest(t)

		// Check if non-existent subscription exists
		exists := subscribeService.HasSubscription("non-existent-id")

		// Assertions
		assert.False(t, exists)
	})
}

func TestSubscribeService_GetSubscriptionTag(t *testing.T) {
	t.Run("Get tag for existing subscription", func(t *testing.T) {
		_, _, subscribeService := setupSubscribeTest(t)

		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}

		// Call Subscribe to create a subscription
		callback := func(subscriptionID string, result interface{}) {}
		subID, _ := subscribeService.Subscribe(testEventTypeLogs, subscribeOptions, callback)

		// Get subscription tag
		tag, exists := subscribeService.GetSubscriptionTag(subID)

		// Assertions
		assert.True(t, exists)
		assert.NotEmpty(t, tag)
	})

	t.Run("Get tag for non-existent subscription", func(t *testing.T) {
		_, _, subscribeService := setupSubscribeTest(t)

		// Get tag for non-existent subscription
		tag, exists := subscribeService.GetSubscriptionTag("non-existent-id")

		// Assertions
		assert.False(t, exists)
		assert.Empty(t, tag)
	})
}

func TestSubscribeService_NotifySubscribers(t *testing.T) {
	t.Run("Notify subscribers for existing tag", func(t *testing.T) {
		_, cacheService, subscribeService := setupSubscribeTest(t)

		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}

		// Create a channel to track callback invocation
		callbackCh := make(chan struct{}, 1)

		// Call Subscribe to create a subscription
		callback := func(subscriptionID string, result interface{}) {
			callbackCh <- struct{}{}
		}

		// Get the tag that will be used
		subID, _ := subscribeService.Subscribe(testEventTypeLogs, subscribeOptions, callback)
		tag, _ := subscribeService.GetSubscriptionTag(subID)

		// Set up the cache to return not found for the notification key
		cacheService.Set(context.Background(), "test-key", false, time.Second)

		// Notify subscribers
		testData := map[string]string{"key": "value"}
		subscribeService.NotifySubscribers(tag, testData)

		// Wait for callback to be called or timeout
		select {
		case <-callbackCh:
			// Callback was called, test passes
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Callback was not called within timeout")
		}
	})

	t.Run("Notify subscribers for non-existent tag", func(t *testing.T) {
		_, _, subscribeService := setupSubscribeTest(t)

		// This should not panic or cause errors
		subscribeService.NotifySubscribers("non-existent-tag", "test-data")
	})
}

func TestCreateSubscriptionTag(t *testing.T) {
	t.Run("Create tag for logs event", func(t *testing.T) {
		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}

		// Create expected tag structure
		expectedTagData := struct {
			Event               string   `json:"event"`
			Address             []string `json:"address,omitempty"`
			Topics              []string `json:"topics,omitempty"`
			IncludeTransactions bool     `json:"includeTransactions,omitempty"`
		}{
			Event:   testEventTypeLogs,
			Address: []string{"0x123"},
			Topics:  []string{"0x456"},
		}
		expectedTagJSON, _ := json.Marshal(expectedTagData)

		// Call the function
		tag, err := service.CreateSubscriptionTag(testEventTypeLogs, subscribeOptions)

		// Assertions
		assert.NoError(t, err)

		// Unmarshal both tags to compare as objects
		var actualTagObj, expectedTagObj map[string]interface{}
		json.Unmarshal([]byte(tag), &actualTagObj)
		json.Unmarshal(expectedTagJSON, &expectedTagObj)

		assert.Equal(t, expectedTagObj, actualTagObj)
	})

	t.Run("Create tag for block event", func(t *testing.T) {
		// Setup test data
		subscribeOptions := domain.SubscribeOptions{
			IncludeTransactions: true,
		}

		// Create expected tag structure
		expectedTagData := struct {
			Event               string   `json:"event"`
			Address             []string `json:"address,omitempty"`
			Topics              []string `json:"topics,omitempty"`
			IncludeTransactions bool     `json:"includeTransactions,omitempty"`
		}{
			Event:               testEventTypeBlock,
			IncludeTransactions: true,
		}
		expectedTagJSON, _ := json.Marshal(expectedTagData)

		// Call the function
		tag, err := service.CreateSubscriptionTag(testEventTypeBlock, subscribeOptions)

		// Assertions
		assert.NoError(t, err)

		// Unmarshal both tags to compare as objects
		var actualTagObj, expectedTagObj map[string]interface{}
		json.Unmarshal([]byte(tag), &actualTagObj)
		json.Unmarshal(expectedTagJSON, &expectedTagObj)

		assert.Equal(t, expectedTagObj, actualTagObj)
	})
}
