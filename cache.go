package cache

// a memory capped LRU cache
import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type Clock struct {
	instant time.Time
}

func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.instant
}

// An Entry represents a value in a Cache
type Entry struct {
	lock        sync.RWMutex
	data        []byte
	expiration  time.Time
	listElement *list.Element
}

// Update will set the value and expiration of an entry in a thread-safe manner
func (e *Entry) Update(value []byte, expiration time.Time) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.data = value
	e.expiration = expiration
}

// Read will return the value of an entry in a thread-safe manner
func (e *Entry) Read() []byte {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.data
}

// A Cache is an optionally memory-capped LRU cache
type Cache struct {
	sync.Mutex
	data            map[string]*Entry
	lru             *list.List
	ticker          *time.Ticker
	bytesReferenced uint64
	clock           *Clock
}

func (c *Cache) Read(key string) []byte {
	c.Lock()

	e, ok := c.data[key]
	if !ok {
		c.Unlock()
		return nil
	}

	if e.expiration.Before(c.clock.Now()) {
		c.lru.Remove(e.listElement)
		delete(c.data, key)
		c.Unlock()
		return nil
	}

	c.lru.MoveToFront(e.listElement)
	c.Unlock()

	return e.Read()
}

// Set performs a thread-safe upsert operation on a cache
func (c *Cache) Set(key string, value []byte, expiration time.Time) {
	c.Lock()
	entry, exists := c.data[key]
	if exists {
		c.markEntryTouched(entry)
		c.Unlock()
		// The item may have been removed from the cache between releasing the lock
		// and performing the write below, this would only happen if entire cache
		// flushed through during update
		entry.Update(value, expiration)
	} else {
		lruElement := c.lru.PushFront(key)

		c.data[key] = &Entry{
			data:        value,
			expiration:  expiration,
			listElement: lruElement,
		}

		c.bytesReferenced += uint64(len(value))

		c.Unlock()
	}
}

// Not thread safe
func (c *Cache) markEntryTouched(e *Entry) {
	if e.listElement != nil {
		c.lru.MoveToFront(e.listElement)
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
			c.Lock()
			for c.bytesReferenced > memoryLimit && c.lru.Len() > 0 {
				entryKey := c.lru.Remove(c.lru.Back()).(string)
				entry := c.data[entryKey]
				c.bytesReferenced -= uint64(len(entry.data))
				delete(c.data, entryKey)
			}
			c.Unlock()
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

// NewCache constructs an optionally memory-capped LRU cache
func NewCache() *Cache {
	cache := &Cache{
		data:            map[string]*Entry{},
		lru:             list.New(),
		bytesReferenced: 0,
	}
	return cache
}
