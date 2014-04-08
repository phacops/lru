package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/dakis/lru"
)

const (
	CACHE_SIZE = 1 << 20
)

func main() {
	cache := lru.New(CACHE_SIZE, "/tmp")
	defer cache.Clear()

	key := "random_bytes"
	randomData := make([]byte, 32*1024)

	_, err := rand.Read(randomData)

	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	cache.Set(key, randomData)

	data, ok := cache.Get(key)

	if ok {
		fmt.Println("data was successfully retrieve from the cache.")

		if bytes.Equal(randomData, data) {
			fmt.Println("data match what was inserted. Yeah!")
		}
	} else {
		fmt.Println("couldn't retrieve the data from the cache.")
	}
}
