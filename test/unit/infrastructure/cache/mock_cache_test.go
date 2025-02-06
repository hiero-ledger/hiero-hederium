package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/LimeChain/Hederium/test/unit/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMockCacheService_SetAndGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var cacheService cache.CacheService = mocks.NewMockCacheService(ctrl)
	ctx := context.Background()

	// Test setting and getting a string value
	testValue := "test_value"
	mockCache := cacheService.(*mocks.MockCacheService)
	mockCache.EXPECT().
		Set(ctx, "test_key", testValue, time.Minute).
		Return(nil)

	var retrievedValue string
	mockCache.EXPECT().
		Get(ctx, "test_key", &retrievedValue).
		DoAndReturn(func(_ context.Context, _ string, out *string) error {
			*out = testValue
			return nil
		})

	err := cacheService.Set(ctx, "test_key", testValue, time.Minute)
	assert.NoError(t, err)

	err = cacheService.Get(ctx, "test_key", &retrievedValue)
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

	mockCache.EXPECT().
		Set(ctx, "test_struct", testStruct, time.Minute).
		Return(nil)

	var retrievedStruct TestStruct
	mockCache.EXPECT().
		Get(ctx, "test_struct", &retrievedStruct).
		DoAndReturn(func(_ context.Context, _ string, out *TestStruct) error {
			*out = testStruct
			return nil
		})

	err = cacheService.Set(ctx, "test_struct", testStruct, time.Minute)
	assert.NoError(t, err)

	err = cacheService.Get(ctx, "test_struct", &retrievedStruct)
	assert.NoError(t, err)
	assert.Equal(t, testStruct, retrievedStruct)
}

func TestMockCacheService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var cacheService cache.CacheService = mocks.NewMockCacheService(ctrl)
	ctx := context.Background()

	testValue := "test_value"
	mockCache := cacheService.(*mocks.MockCacheService)
	mockCache.EXPECT().
		Set(ctx, "test_key", testValue, time.Minute).
		Return(nil)

	mockCache.EXPECT().
		Delete(ctx, "test_key").
		Return(nil)

	var retrievedValue string
	mockCache.EXPECT().
		Get(ctx, "test_key", &retrievedValue).
		Return(errors.New("key not found"))

	err := cacheService.Set(ctx, "test_key", testValue, time.Minute)
	assert.NoError(t, err)

	err = cacheService.Delete(ctx, "test_key")
	assert.NoError(t, err)

	err = cacheService.Get(ctx, "test_key", &retrievedValue)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestMockCacheService_GetNonExistent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var cacheService cache.CacheService = mocks.NewMockCacheService(ctrl)
	ctx := context.Background()

	var value string
	mockCache := cacheService.(*mocks.MockCacheService)
	mockCache.EXPECT().
		Get(ctx, "non_existent_key", &value).
		Return(errors.New("key not found"))

	err := cacheService.Get(ctx, "non_existent_key", &value)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}
