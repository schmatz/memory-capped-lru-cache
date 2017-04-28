# Memory Capped LRU Cache
![CircleCI](https://circleci.com/gh/schmatz/memory-capped-lru-cache.svg?style=shield&circle-token=b1f1b20c9eabca6b95b5fa3a73fdf8bf4b592619)
[![](https://godoc.org/github.comschmatz/memory-capped-lru-cache?status.svg)](http://godoc.org/github.com/schmatz/memory-capped-lru-cache)

This package provides an LRU `string -> []byte` cache with a few nice features:
* Support for evicting the least recently used items to reduce the amount of data referenced by the cache
  * Both synchronous and asynchronous eviction supported
* Setting TTL on items with lazy eviction
* Good performance
  * 176 ns/insert over 10M inserts with 8 4GHz cores contending (`BenchmarkConcurrentInserts`)

This cache is ideal for items of roughly uniform size with a TTL.
