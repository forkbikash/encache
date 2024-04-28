# encache

[![Go Reference](https://pkg.go.dev/badge/github.com/forkbikash/encache.svg)](https://pkg.go.dev/github.com/forkbikash/encache)
[![Go Report Card](https://goreportcard.com/badge/github.com/forkbikash/encache)](https://goreportcard.com/report/github.com/forkbikash/encache)

`encache` is a Go package that provides a caching mechanism for function calls. It allows you to cache the results of expensive function calls and retrieve them from the cache instead of recomputing them. This can greatly improve the performance of your application, especially when dealing with computationally intensive or I/O-bound operations.

## Features

- Support for in-memory and Redis caching
- Automatic cache expiration and periodic expiration of stale entries
- Locking mechanisms to ensure thread-safety
- Customizable cache key generation
- Option to cache function results even when errors occur

## Installation

```bash
go get github.com/forkbikash/encache
```

## Usage

Here's a simple example of how to use the encache package:

```go
package main

import (
	"fmt"
	"time"

	"github.com/forkbikash/encache"
)

func expensiveOperation(a, b int) (int, error) {
	// Simulate an expensive operation
	time.Sleep(2 * time.Second)
	return a + b, nil
}

func main() {
	// Create a new in-memory cache implementation
	mapCache := encache.NewMapCacheImpl()
	cacheKeyImpl := encache.NewDefaultCacheKeyImpl()
	lockImpl := encache.NewMuLockImpl()

	// Create a new encache instance
	encache := encache.NewEncache(lockImpl, mapCache, cacheKeyImpl, false)

	// Wrap the expensive function with caching
	cachedExpensiveOperation := encache.CachedFunc(expensiveOperation, time.Minute)

	// Call the cached function
	result, err := cachedExpensiveOperation(2, 3)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Result:", result)

	// Subsequent calls will retrieve the result from the cache
	result, err = cachedExpensiveOperation(2, 3)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Result (cached):", result)
}
```

This example demonstrates how to create a new encache instance with an in-memory cache (MapCacheImpl), a default cache key implementation (DefaultCacheKeyImpl), and a mutex-based lock implementation (MuLockImpl). It then wraps the expensiveOperation function with the CachedFunc function, which returns a new function that will cache the results of expensiveOperation.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have any improvements, bug fixes, or new features to propose.

## Future developments

- Cache invalidation strategies: Aside from the simple expiration-based cache invalidation, we can add support for other strategies like manual invalidation, LRU (Least Recently Used) eviction, or event-based invalidation (e.g., invalidating the cache when the underlying data changes).
- Monitoring and metrics: Provide metrics and monitoring capabilities to help users understand the cache's performance, hit/miss rates, and other relevant statistics.
- Adaptive caching: Implement an adaptive caching mechanism that can automatically adjust the cache size, eviction policy, or other parameters based on the workload and usage patterns.
- Asynchronous cache updates: Provide an asynchronous cache update mechanism to allow for non-blocking cache population and update operations.
- Change package structure

```dir
caching/
├── backend/
│   ├── memory/
│   ├── redis/
│   └── memcached/
├── policy/
│   ├── expiration/
│   ├── lru/
│   └── invalidation/
├── serializer/
│   ├── json/
│   ├── gob/
│   └── msgpack/
├── cache.go
├── options.go
├── metrics.go
└── utils.go
```
