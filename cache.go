package cache

// a memory capped LRU cache
import (
	"container/list"
	"errors"
	"sync"
	"time"
)

// A Cache is an optionally memory-capped LRU cache
type Cache struct {
	sync.Mutex
	data            map[string]*entry
	lru             *list.List
	ticker          *time.Ticker
	bytesReferenced uint64
	clock           *clock
}

// NewCache constructs an optionally memory-capped LRU cache
func NewCache() *Cache {
	cache := &Cache{
		data:            map[string]*entry{},
		lru:             list.New(),
		bytesReferenced: 0,
	}
	return cache
}

// Get will return a value from the cache and nil if
// the item is nonexistent or expired.
func (c *Cache) Get(key string) []byte {
	c.Lock()
	defer c.Unlock()

	e, ok := c.data[key]
	if !ok {
		return nil
	}

	if e.expiration.Before(c.clock.now()) {
		c.lru.Remove(e.listElement)
		delete(c.data, key)
		return nil
	}

	c.lru.MoveToFront(e.listElement)
	return e.read()
}

// Set performs an upsert operation on a cache
func (c *Cache) Set(key string, value []byte, expiration time.Time) {
	c.Lock()
	defer c.Unlock()

	e, exists := c.data[key]
	if exists {
		c.lru.MoveToFront(e.listElement)
		// The item may have been removed from the cache between releasing the lock
		// and performing the write below, this would only happen if entire cache
		// flushed through during update
		e.update(value, expiration)
	} else {
		lruElement := c.lru.PushFront(key)

		c.data[key] = &entry{
			data:        value,
			expiration:  expiration,
			listElement: lruElement,
		}

		c.bytesReferenced += uint64(len(value))
	}
}

// StartEviction starts the eviction process or returns an error if one exists
func (c *Cache) StartEviction(memoryLimit uint64, checkInterval time.Duration) error {
	c.Lock()
	defer c.Unlock()

	if c.ticker != nil {
		return errors.New("Eviction has started, you must stopEviction")
	}

	c.ticker = time.NewTicker(checkInterval)
	go func() {
		for _ = range c.ticker.C {
			c.ShrinkCache(memoryLimit)
		}
	}()

	return nil
}

// StopEviction will halt any background eviction process if it exists
func (c *Cache) StopEviction() {
	c.Lock()
	defer c.Unlock()

	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}
}

// BytesReferenced will return the sum of the size of all items referenced
// by the cache
func (c *Cache) BytesReferenced() uint64 {
	c.Lock()
	defer c.Unlock()

	return c.bytesReferenced
}

// ShrinkCache will delete the least recently used items until the
// number of bytes referenced by the cache falls below the target
func (c *Cache) ShrinkCache(targetReferencedBytes uint64) {
	c.Lock()
	for c.bytesReferenced > targetReferencedBytes && c.lru.Len() > 0 {
		entryKey := c.lru.Remove(c.lru.Back()).(string)
		entry := c.data[entryKey]
		c.bytesReferenced -= uint64(len(entry.data))
		delete(c.data, entryKey)
	}
	c.Unlock()
}
