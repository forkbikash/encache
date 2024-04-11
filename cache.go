package cacheme

import (
	"log"
	"reflect"
	"time"
)

// do these first:
// way to return error if some error in MakeFunc
// learn the differences between interface{}, kind, type
// write unit tests
// write doc

// features(implement all of them):
// Cache invalidation strategies: Aside from the simple expiration-based cache invalidation, we can add support for other strategies like manual invalidation, LRU (Least Recently Used) eviction, or event-based invalidation (e.g., invalidating the cache when the underlying data changes).
// Monitoring and metrics: Provide metrics and monitoring capabilities to help users understand the cache's performance, hit/miss rates, and other relevant statistics.
// Adaptive caching: Implement an adaptive caching mechanism that can automatically adjust the cache size, eviction policy, or other parameters based on the workload and usage patterns.
// Asynchronous cache updates: Provide an asynchronous cache update mechanism to allow for non-blocking cache population and update operations.
// package structure
// caching/
// ├── backend/
// │   ├── memory/
// │   ├── redis/
// │   └── memcached/
// ├── policy/
// │   ├── expiration/
// │   ├── lru/
// │   └── invalidation/
// ├── serializer/
// │   ├── json/
// │   ├── gob/
// │   └── msgpack/
// ├── cache.go
// ├── options.go
// ├── metrics.go
// └── utils.go

type config struct {
	LockImpl      LockType
	CacheImpl     CacheType
	CacheKeyImpl  CacheKeyType
	SetCacheOnErr bool
}

func NewConfig(LockImpl LockType, CacheImpl CacheType, CacheKeyImpl CacheKeyType, setCacheOnErr bool) *config {
	return &config{
		LockImpl:      LockImpl,
		CacheImpl:     CacheImpl,
		CacheKeyImpl:  CacheKeyImpl,
		SetCacheOnErr: setCacheOnErr,
	}
}

type CacheType interface {
	Get(string, reflect.Type) ([]reflect.Value, bool, error)
	Set(string, []reflect.Value, time.Duration) error
	Serialize([]reflect.Value) (string, error)
	Deserialize(string, reflect.Type) ([]reflect.Value, error)
	Expire(time.Duration)
}

type CacheKeyType interface {
	Key([]reflect.Value) string
}

type LockType interface {
	lock(...string) error
	unlock(...string) error
}

// closure = returned anonymous inner function + outer context(variables defined outside of inner function)
// func CachedFunc[T any](f T, lockImpl LockType, cacheImpl CacheType, cacheKeyImpl CacheKeyType, expiry time.Duration) T {
func CachedFunc[T any](f T, config config, expiry time.Duration) T {
	fValue := reflect.ValueOf(f)
	fType := fValue.Type()

	if fType.Kind() != reflect.Func {
		panic("input is not a function")
	}

	config.CacheImpl.Expire(expiry)

	return reflect.MakeFunc(fType, func(args []reflect.Value) []reflect.Value {
		key := config.CacheKeyImpl.Key(args)

		lockerr := config.LockImpl.lock()
		if lockerr != nil {
			log.Println("error in lock: ", lockerr)
			return callAndSet(fValue, args, config.CacheImpl, config.SetCacheOnErr, key, expiry)
		}
		defer func() {
			unlockerr := config.LockImpl.unlock()
			if unlockerr != nil {
				log.Println("error in unlock: ", unlockerr)
			}
		}()

		getres, found, geterr := config.CacheImpl.Get(key, fType)
		if geterr != nil {
			log.Println("error in get: ", geterr)
			return callAndSet(fValue, args, config.CacheImpl, config.SetCacheOnErr, key, expiry)
		}
		if found {
			return getres
		}

		return callAndSet(fValue, args, config.CacheImpl, config.SetCacheOnErr, key, expiry)
	}).Interface().(T)
}

func callAndSet(fValue reflect.Value, args []reflect.Value, cacheImpl CacheType, setCacheOnErr bool, key string, expiry time.Duration) []reflect.Value {
	callres := fValue.Call(args)

	var callreserr error
	for _, v := range callres {
		if v.Kind() == reflect.Interface {
			if errVal, ok := v.Interface().(error); ok {
				callreserr = errVal
				break
			}
		}
	}

	if setCacheOnErr || callreserr == nil {
		seterr := cacheImpl.Set(key, callres, expiry)
		if seterr != nil {
			log.Println("error in set: ", seterr)
		}
	}

	return callres
}
