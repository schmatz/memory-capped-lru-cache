package cache

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestSetGetUpdate(t *testing.T) {
	cache := NewCache()

	testKey := "test"
	testVal := []byte("This is a test!")
	expiration := time.Now().Add(time.Hour)

	cache.Set(testKey, testVal, expiration)

	readVal := cache.Get(testKey)
	if !reflect.DeepEqual(readVal, testVal) {
		t.Error("Key put in store not equal to key read from store")
	}

	testVal = []byte("Different value")
	cache.Set(testKey, testVal, expiration)

	readVal = cache.Get(testKey)
	if !reflect.DeepEqual(readVal, testVal) {
		t.Error("Key put in store not equal to key read from store")
	}
}

func TestNonexistentGet(t *testing.T) {
	cache := NewCache()

	nonexistent := cache.Get("lol")
	if nonexistent != nil {
		t.Error("Non-existent values should be nil")
	}
}

func TestEviction(t *testing.T) {
	cache := NewCache()

	testKey := "test"
	testVal := []byte("This is a test!")
	expiration := time.Now().Add(time.Hour)

	cache.Set(testKey, testVal, expiration)

	// This is pretty jank, fix this testing by not relying on time
	cache.StartEviction(0, time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	cache.StopEviction()

	shouldBeEvicted := cache.Get(testKey)
	if shouldBeEvicted != nil {
		t.Error("Expected the value to be evicted")
	}
}

func TestStartEvictionTwice(t *testing.T) {
	cache := NewCache()
	err := cache.StartEviction(5000, time.Second)
	if err != nil {
		t.Error("Expected no error when starting eviction for the first time")
	}

	err = cache.StartEviction(5000, time.Second)
	if err == nil {
		t.Error("Expected error when starting eviction twice")
	}
}

func TestExpiration(t *testing.T) {
	cache := NewCache()

	testKey := "test"
	testVal := []byte("This is a test!")
	expiration := time.Now().Add(time.Hour)

	cache.Set(testKey, testVal, expiration)

	cache.clock = &clock{instant: expiration.Add(1 * time.Second)}

	shouldBeEvicted := cache.Get(testKey)

	if shouldBeEvicted != nil {
		t.Error("Expected the expired item to be evicted")
	}
}

func TestBytesReferenced(t *testing.T) {
	cache := NewCache()

	testKey := "test"
	testVal := []byte("This is a test!")
	expiration := time.Now().Add(time.Hour)

	cache.Set(testKey, testVal, expiration)

	size := cache.BytesReferenced()
	if size != uint64(len(testVal)) {
		t.Error("Expected size of cache to be equal to sum of items")
	}
}

func BenchmarkConcurrentInserts(b *testing.B) {
	cache := NewCache()
	expiration := time.Now().Add(time.Hour)
	var strings []string
	for i := 0; i < 100000; i++ {
		strings = append(strings, strconv.Itoa(i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			s := strings[counter%len(strings)]
			cache.Set(s, []byte(s), expiration)
			counter++
		}
	})
}
