package encache

import (
	"log"
	"reflect"
	"time"
)

type Encache struct {
	LockImpl      LockType
	CacheImpl     CacheType
	CacheKeyImpl  CacheKeyType
	SetCacheOnErr bool
}

func NewEncache(LockImpl LockType, CacheImpl CacheType, CacheKeyImpl CacheKeyType, setCacheOnErr bool, staleRemovalPeriod time.Duration) Encache {
	encache := Encache{
		LockImpl:      LockImpl,
		CacheImpl:     CacheImpl,
		CacheKeyImpl:  CacheKeyImpl,
		SetCacheOnErr: setCacheOnErr,
	}

	encache.CacheImpl.PeriodicExpire(staleRemovalPeriod)

	return encache
}

type CacheType interface {
	Get(string, reflect.Type) ([]reflect.Value, bool, error)
	Set(string, []reflect.Value, time.Duration) error
	Serialize([]reflect.Value) (string, error)
	Deserialize(string, reflect.Type) ([]reflect.Value, error)
	PeriodicExpire(time.Duration)
	Expire(string, time.Duration) error
}

type CacheKeyType interface {
	Key([]reflect.Value) string
}

type LockType interface {
	lock(...string) error
	unlock(...string) error
}

// closure = returned anonymous inner function + outer context(variables defined outside of inner function)
// get the instance of function f with caching logic
func CachedFunc[T any](f T, encache Encache, expiry time.Duration) T {
	fValue := reflect.ValueOf(f)
	fType := fValue.Type()

	if fType.Kind() != reflect.Func {
		panic("input is not a function")
	}

	return reflect.MakeFunc(fType, func(args []reflect.Value) []reflect.Value {
		key := encache.CacheKeyImpl.Key(args)

		lockerr := encache.LockImpl.lock()
		if lockerr != nil {
			log.Println("error in lock: ", lockerr)
			return callAndSet(fValue, args, encache.CacheImpl, encache.SetCacheOnErr, key, expiry)
		}
		defer func() {
			unlockerr := encache.LockImpl.unlock()
			if unlockerr != nil {
				log.Println("error in unlock: ", unlockerr)
			}
		}()

		getres, found, geterr := encache.CacheImpl.Get(key, fType)
		if geterr != nil {
			log.Println("error in get: ", geterr)
			return callAndSet(fValue, args, encache.CacheImpl, encache.SetCacheOnErr, key, expiry)
		}
		if found {
			log.Println("cache found")
			return getres
		}

		log.Println("cache not found")
		return callAndSet(fValue, args, encache.CacheImpl, encache.SetCacheOnErr, key, expiry)
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

// change expiration or expire immediately
// to expire immediately, pass 0 as expiry
func Expire(encache Encache, key string, expiry time.Duration) error {
	return encache.CacheImpl.Expire(key, expiry)
}
