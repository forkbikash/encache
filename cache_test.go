package cacheme

import (
	"fmt"
	"time"
)

func Example() {
	add := func(a, b int) (int, int, error) {
		return a + b, 3, nil
	}

	// cachedAdd := CachedFunc(add, NewMuLockImpl())
	// cachedAdd := CachedFunc(add, NewRedisLockImpl(redis.NewClient(nil)), NewMapCacheImpl(time.Second), NewDefaultCacheKeyImpl(), 5*time.Second)

	config := NewConfig(NewNoLockImpl(), NewMapCacheImpl(), NewDefaultCacheKeyImpl(), false)
	cachedAdd := CachedFunc(add, *config, 5*time.Second)

	fmt.Println(cachedAdd(2, 3)) // Output: 5
	fmt.Println(cachedAdd(2, 3)) // Output: 5 (cached)
	fmt.Println(cachedAdd(4, 5)) // Output: 9
	fmt.Println(cachedAdd(4, 5)) // Output: 9 (cached)
}
