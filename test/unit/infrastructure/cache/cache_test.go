package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*zap.Logger, cache.CacheService) {
	logger, _ := zap.NewDevelopment()
	cacheService := cache.NewMemoryCache(time.Minute*5, time.Minute)
	return logger, cacheService
}

func TestNewMemoryCache(t *testing.T) {
	memCache := cache.NewMemoryCache(time.Minute*5, time.Minute)
	assert.NotNil(t, memCache)
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	memCache := cache.NewMemoryCache(time.Minute*5, time.Minute)
	ctx := context.Background()

	// Test setting and getting a string value
	testValue := "test_value"
	err := memCache.Set(ctx, "test_key", testValue, time.Minute)
	assert.NoError(t, err)

	var retrievedValue string
	err = memCache.Get(ctx, "test_key", &retrievedValue)
	assert.NoError(t, err)
	assert.Equal(t, testValue, retrievedValue)

	// Test setting and getting a struct value
	type TestStruct struct {
		Field1 string
		Field2 int
	}
	testStruct := TestStruct{
		Field1: "test",
		Field2: 123,
	}

	err = memCache.Set(ctx, "test_struct", testStruct, time.Minute)
	assert.NoError(t, err)

	var retrievedStruct TestStruct
	err = memCache.Get(ctx, "test_struct", &retrievedStruct)
	assert.NoError(t, err)
	assert.Equal(t, testStruct, retrievedStruct)
}

func TestMemoryCache_Delete(t *testing.T) {
	memCache := cache.NewMemoryCache(time.Minute*5, time.Minute)
	ctx := context.Background()

	// Set a value
	testValue := "test_value"
	err := memCache.Set(ctx, "test_key", testValue, time.Minute)
	assert.NoError(t, err)

	// Delete the value
	err = memCache.Delete(ctx, "test_key")
	assert.NoError(t, err)

	// Try to get the deleted value
	var retrievedValue string
	err = memCache.Get(ctx, "test_key", &retrievedValue)
	assert.Error(t, err)
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	memCache := cache.NewMemoryCache(time.Minute*5, time.Minute)
	ctx := context.Background()

	var value string
	err := memCache.Get(ctx, "non_existent_key", &value)
	assert.Error(t, err)
}

func TestMemoryCache_Expiration(t *testing.T) {
	memCache := cache.NewMemoryCache(time.Minute*5, time.Minute)
	ctx := context.Background()

	// Set a value with a very short expiration
	testValue := "test_value"
	err := memCache.Set(ctx, "test_key", testValue, time.Millisecond*100)
	assert.NoError(t, err)

	// Wait for the value to expire
	time.Sleep(time.Millisecond * 200)

	// Try to get the expired value
	var retrievedValue string
	err = memCache.Get(ctx, "test_key", &retrievedValue)
	assert.Error(t, err)
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	memCache := cache.NewMemoryCache(time.Minute*5, time.Minute)
	ctx := context.Background()

	// Test concurrent access to the cache
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("key_%d", index)
			value := fmt.Sprintf("value_%d", index)

			err := memCache.Set(ctx, key, value, time.Minute)
			assert.NoError(t, err)

			var retrievedValue string
			err = memCache.Get(ctx, key, &retrievedValue)
			assert.NoError(t, err)
			assert.Equal(t, value, retrievedValue)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
