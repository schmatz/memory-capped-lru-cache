package cache

import "testing"
import "time"
import "reflect"

func TestSetReadUpdate(t *testing.T) {
	cache := NewCache()

	testKey := "test"
	testVal := []byte("This is a test!")
	expiration := time.Now().Add(time.Hour)

	cache.Set(testKey, testVal, expiration)

	readVal := cache.Read(testKey)
	if !reflect.DeepEqual(readVal, testVal) {
		t.Error("Key put in store not equal to key read from store")
	}

	testVal = []byte("Different value")
	cache.Set(testKey, testVal, expiration)

	readVal = cache.Read(testKey)
	if !reflect.DeepEqual(readVal, testVal) {
		t.Error("Key put in store not equal to key read from store")
	}
}

func TestNonexistentRead(t *testing.T) {
	cache := NewCache()

	nonexistent := cache.Read("lol")
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

	shouldBeEvicted := cache.Read(testKey)
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
