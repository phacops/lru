package lru

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestInitialState(t *testing.T) {
	cache := New(1, "/tmp")
	defer cache.Clear()

	if currentSize := cache.Size(); currentSize != 0 {
		t.Errorf("size = %v, want 0", currentSize)
	}

	if maxSize := cache.MaxSize(); maxSize != 1 {
		t.Errorf("maxSize = %v, want 1", maxSize)
	}
}

func TestSetInsertsValue(t *testing.T) {
	key := "lru.go"
	data, err := ioutil.ReadFile(key)

	if err != nil {
		t.Errorf("Couldn't read %v", key)
	}

	size := uint64(len(data))
	cache := New(size, "/tmp")
	defer cache.Clear()

	cache.Set(key, data)

	value, ok := cache.Get(key)

	if !ok || !bytes.Equal(value, data) {
		t.Errorf("Cache has incorrect value")
	}
}

func TestSetUpdatesSize(t *testing.T) {
	fileName := "lru.go"
	someValue, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Errorf("Couldn't read %v", fileName)
	}

	fileSize := uint64(len(someValue))
	cache := New(fileSize, "/tmp")
	defer cache.Clear()

	emptyValue := []byte{}
	key := "key1"

	cache.Set(key, emptyValue)

	if currentSize := cache.Size(); currentSize != 0 {
		t.Errorf("cache.Size() = %v, expected 0", currentSize)
	}

	cache.Set(fileName, someValue)

	if currentSize := cache.Size(); currentSize != fileSize {
		t.Errorf("cache.Size() = %v, expected %v", currentSize, fileSize)
	}
}

func TestGetNonExistent(t *testing.T) {
	cache := New(1, "/tmp")

	if _, ok := cache.Get("i don't exist"); ok {
		t.Error("Cache returned a crap value after no inserts.")
	}
}

func TestDelete(t *testing.T) {
	key := "lru.go"
	value, err := ioutil.ReadFile(key)

	if err != nil {
		t.Errorf("Couldn't read %v", key)
	}

	size := uint64(len(value))
	cache := New(size, "/tmp")
	defer cache.Clear()

	if cache.Delete(key) {
		t.Error("Item unexpectedly already in cache.")
	}

	cache.Set(key, value)

	if !cache.Delete(key) {
		t.Error("Expected item to be in cache.")
	}

	if currentSize := cache.Size(); currentSize != 0 {
		t.Errorf("cache.Size() = %v, expected 0", currentSize)
	}

	if _, ok := cache.Get(key); ok {
		t.Error("Cache returned a value after deletion.")
	}
}

func TestClear(t *testing.T) {
	key := "lru.go"
	value, err := ioutil.ReadFile(key)

	if err != nil {
		t.Errorf("Couldn't read %v", key)
	}

	cache := New(1, "/tmp")
	defer cache.Clear()

	cache.Set(key, value)
	cache.Clear()

	if currentSize := cache.Size(); currentSize != 0 {
		t.Errorf("cache.Size() = %v, expected 0 after Clear()", currentSize)
	}
}

func TestCapacityIsObeyed(t *testing.T) {
	key := "lru.go"
	value, err := ioutil.ReadFile(key)

	if err != nil {
		t.Errorf("Couldn't read %v", key)
	}

	size := uint64(len(value)) * 3
	cache := New(size, "/tmp")
	defer cache.Clear()

	cache.Set("key1", value)
	cache.Set("key2", value)
	cache.Set("key3", value)

	if currentSize := cache.Size(); currentSize != size {
		t.Errorf("cache.Size() = %v, expected %v", currentSize, size)
	}

	cache.Set("key4", value)

	if currentSize := cache.Size(); currentSize != size {
		t.Errorf("post-evict cache.Size() = %v, expected %v", currentSize, size)
	}
}

func TestLRUIsEvicted(t *testing.T) {
	key := "lru.go"
	value, err := ioutil.ReadFile(key)

	if err != nil {
		t.Errorf("Couldn't read %v", key)
	}

	size := uint64(len(value)) * 3
	cache := New(size, "/tmp")
	defer cache.Clear()

	cache.Set("key1", value)
	cache.Set("key2", value)
	cache.Set("key3", value)

	cache.Get("key3")
	cache.Get("key2")
	cache.Get("key1")

	cache.Set("key0", value)

	if _, ok := cache.Get("key3"); ok {
		t.Error("Least recently used element was not evicted.")
	}
}
