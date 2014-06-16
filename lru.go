package lru

import (
	"container/list"
	"encoding/base64"
	"fmt"
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

	list  *list.List
	table map[string]*list.Element
	size  uint64

	maxSize uint64
	path    string
}

func hashCacheKey(data string) string {
	hash := fnv.New64a()
	hash.Write([]byte(data))

	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func New(maxSize uint64, path string) *Cache {
	return &Cache{
		list:    list.New(),
		table:   make(map[string]*list.Element),
		maxSize: maxSize,
		path:    path,
	}
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

	cache.size -= element.Value.(*object).size

	return true
}

func (cache *Cache) Clear() {
	cache.Lock()
	defer cache.Unlock()

	cache.clearFiles()
	cache.list.Init()
	cache.table = make(map[string]*list.Element)
	cache.size = 0
}

func (cache *Cache) Size() uint64 {
	cache.Lock()
	defer cache.Unlock()

	return cache.size
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

func (cache *Cache) keys() []string {
	keys := make([]string, 0, cache.list.Len())

	for element := cache.list.Front(); element != nil; element = element.Next() {
		keys = append(keys, element.Value.(*object).key)
	}

	return keys
}

func (cache *Cache) moveToFront(element *list.Element) {
	cache.list.MoveToFront(element)
	element.Value.(*object).accessTime = time.Now()
}

func (cache *Cache) addNew(key string, value []byte) {
	size := uint64(len(value))

	fmt.Printf("lru: new object of size %d\n", size)

	if size > cache.maxSize {
		fmt.Println("lru: file is too large")
		return
	}

	newObject := &object{key, size, time.Now()}

	cache.trim(cache.size + newObject.size)

	if _, err := os.Stat(cache.FilePath(key)); os.IsNotExist(err) {
		err := ioutil.WriteFile(cache.FilePath(key), value, 0644)

		if err != nil {
			fmt.Println("lru: " + err.Error())
			return
		}

		element := cache.list.PushFront(newObject)
		cache.table[key] = element
		cache.size += (*newObject).size
		fmt.Printf("lru: added %d, new size is %d\n", (*newObject).size, cache.size)
	} else {
		fmt.Println("lru: file already exist")
	}
}

func (cache *Cache) trim(futureSize uint64) {
	for futureSize > cache.maxSize {
		element := cache.list.Back()

		if element == nil {
			fmt.Println("lru: file is too large")
			return
		}

		value := cache.list.Remove(element).(*object)

		if err := os.RemoveAll(cache.FilePath(value.key)); err != nil {
			fmt.Printf("lru: couldn't delete %s\n", cache.FilePath(value.key))
		}

		delete(cache.table, value.key)

		cache.size -= value.size
		futureSize -= value.size
	}
}

func (cache *Cache) clearFiles() {
	for _, key := range cache.keys() {
		os.RemoveAll(cache.FilePath(key))
	}
}
