package lru

import (
	"container/list"
	"encoding/base64"
	"hash/fnv"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type object struct {
	key        string
	size       uint64
	accessTime time.Time
}

type Cache struct {
	sync.Mutex

	list        *list.List
	table       map[string]*list.Element
	currentSize uint64

	maxSize uint64
	path    string
}

func hashCacheKey(data string) string {
	hash := fnv.New64a()
	hash.Write([]byte(data))

	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func New(maxSize uint64, path string) *Cache {
	cache := Cache{
		list:    list.New(),
		table:   make(map[string]*list.Element),
		maxSize: maxSize,
		path:    path,
	}

	cache.Clear()

	return &cache
}

func (cache *Cache) FilePath(key string) string {
	return cache.path + "/" + hashCacheKey(key)
}

func (cache *Cache) Get(key string) ([]byte, bool) {
	cache.Lock()
	defer cache.Unlock()

	element := cache.table[key]

	if element == nil {
		return nil, false
	}

	cache.moveToFront(element)

	value, err := ioutil.ReadFile(cache.FilePath(element.Value.(*object).key))

	if err != nil {
		return nil, false
	}

	return value, true
}

func (cache *Cache) Set(key string, value []byte) {
	cache.Lock()
	defer cache.Unlock()

	if element := cache.table[key]; element != nil {
		cache.moveToFront(element)
	} else {
		cache.addNew(key, value)
	}
}

func (cache *Cache) Delete(key string) bool {
	cache.Lock()
	defer cache.Unlock()

	element := cache.table[key]

	if element == nil {
		return false
	}

	err := os.Remove(cache.FilePath(key))

	if err != nil {
		return false
	}

	cache.list.Remove(element)
	delete(cache.table, key)

	cache.currentSize -= element.Value.(*object).size

	return true
}

func (cache *Cache) Clear() {
	cache.Lock()
	defer cache.Unlock()

	cache.clearFiles()
	cache.list.Init()
	cache.table = make(map[string]*list.Element)
	cache.currentSize = 0
}

func (cache *Cache) CurrentSize() uint64 {
	cache.Lock()
	defer cache.Unlock()

	return cache.currentSize
}

func (cache *Cache) MaxSize() uint64 {
	cache.Lock()
	defer cache.Unlock()

	return cache.maxSize
}

func (cache *Cache) Oldest() (oldest time.Time) {
	cache.Lock()
	defer cache.Unlock()

	if lastElem := cache.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*object).accessTime
	}

	return
}

func (cache *Cache) moveToFront(element *list.Element) {
	cache.list.MoveToFront(element)
	element.Value.(*object).accessTime = time.Now()
}

func (cache *Cache) addNew(key string, value []byte) {
	newobject := &object{key, uint64(len(value)), time.Now()}

	if _, err := os.Stat(cache.FilePath(key)); os.IsNotExist(err) {
		err := ioutil.WriteFile(cache.FilePath(key), value, 0644)

		if err != nil {
			return
		}
	}

	element := cache.list.PushFront(newobject)
	cache.table[key] = element
	cache.currentSize += newobject.size

	cache.trim()
}

func (cache *Cache) trim() {
	for cache.currentSize > cache.maxSize {
		element := cache.list.Back()

		if element == nil {
			return
		}

		value := cache.list.Remove(element).(*object)

		os.RemoveAll(cache.FilePath(value.key))
		delete(cache.table, value.key)

		cache.currentSize -= value.size
	}
}

func (cache *Cache) clearFiles() {
	files, _ := ioutil.ReadDir(cache.path)

	for _, file := range files {
		os.RemoveAll(cache.FilePath(file.Name()))
	}
}
