package cacheme

import (
	"fmt"
	"time"
)

type addStruct struct{}

func (addStruct) add(a, b int) (int, int, error) {
	return a + b, 3, nil
}

func Example() {
	// cachedAdd := CachedFunc(add, NewMuLockImpl())
	// cachedAdd := CachedFunc(add, NewRedisLockImpl(redis.NewClient(nil)), NewMapCacheImpl(time.Second), NewDefaultCacheKeyImpl(), 5*time.Second)

	encache := NewEncache(NewNoLockImpl(), NewMapCacheImpl(), NewDefaultCacheKeyImpl(), false)
	s := addStruct{}
	cachedAdd := CachedFunc(s.add, *encache, 5*time.Second)

	fmt.Println(cachedAdd(2, 3)) // Output: 5
	fmt.Println(cachedAdd(2, 3)) // Output: 5 (cached)
	fmt.Println(cachedAdd(4, 5)) // Output: 9
	fmt.Println(cachedAdd(4, 5)) // Output: 9 (cached)
}
