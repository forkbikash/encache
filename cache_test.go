package encache

import (
	"errors"
	"testing"
	"time"
)

func TestMapImplCachedFunc(t *testing.T) {
	// Test with MapCacheImpl
	mapCache := NewMapCacheImpl()
	cacheKeyImpl := NewDefaultCacheKeyImpl()
	lockImpl := NewMuLockImpl()

	// Test a simple function
	simpleFunc := func(a, b int) (int, error) {
		return a + b, nil
	}

	cachedSimpleFunc := CachedFunc(simpleFunc, NewEncache(lockImpl, mapCache, cacheKeyImpl, false, time.Second), 5*time.Second)

	result, err := cachedSimpleFunc(2, 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}

	// Test caching
	result, err = cachedSimpleFunc(2, 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}

	// Test expiration
	time.Sleep(5*time.Second + time.Second)
	result, err = cachedSimpleFunc(2, 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}

	// Test function with errors
	errorFunc := func(a, b int) (int, error) {
		if a == 0 {
			return 0, errors.New("division by zero")
		}
		return b / a, nil
	}

	// Cache on error
	cachedErrorFunc := CachedFunc(errorFunc, NewEncache(lockImpl, mapCache, cacheKeyImpl, true, time.Second), 5*time.Second)

	_, err = cachedErrorFunc(0, 10)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Test caching
	result, err = cachedErrorFunc(2, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}

	// Don't cache on error
	cachedErrorFunc = CachedFunc(errorFunc, NewEncache(lockImpl, mapCache, cacheKeyImpl, false, time.Second), 5*time.Second)

	_, err = cachedErrorFunc(0, 10)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Test non caching
	result, err = cachedErrorFunc(2, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}
}
