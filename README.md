lru
===

This library is an easy way to manage a cache to store bytes on disk.

For now, this cache:
+ doesn't allow updates (no update implemented yet)
+ is never going over the limit (we check before writing to the disk instead of triming after)
+ is supposed to be empty at initialization (we do not delete anything, but a 2 keys on 2 differents instances will have the same file name and so value will be the same as the first one written to disk)
+ is goroutine safe (it has a lock mechanism)

Installation
------------
```
go get github.com/dakis/lru
```

Initialize the cache
--------------
```
cache := lru.New(CACHE_SIZE, "/tmp")
defer cache.Clear()
```

Get a value
-----------
```
cache := lru.New(CACHE_SIZE, "/tmp")
defer cache.Clear()

if data, ok := cache.Get("key); ok {
    fmt.Println("value was retrieved")
} else {
    fmt.Println("value doesn't exist")
}
```

Set a value
-----------
```
cache := lru.New(CACHE_SIZE, "/tmp")
defer cache.Clear()

cache.Set(, "/tmp")
```

Delete a value
--------------
```
cache := lru.New(CACHE_SIZE, "/tmp")
defer cache.Clear()

cache.Set("key", []byte("test value"))

if cache.Delete("key") {
    fmt.Println("value was delete")
}
```

Full examples in https://github.com/dakis/lru/tree/master/examples.
