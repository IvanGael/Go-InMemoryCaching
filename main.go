package main

import (
	"fmt"
	"sync"
	"time"
)

// CacheItem represents a single cache item
type CacheItem struct {
	value      interface{}
	expiration int64
}

// Cache represents the in-memory cache
type Cache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]*CacheItem),
	}
	go cache.startEviction()
	return cache
}

// Set adds a new item to the cache with an optional expiration duration
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}
	c.mu.Lock()
	c.items[key] = &CacheItem{
		value:      value,
		expiration: expiration,
	}
	c.mu.Unlock()
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()
	if !found {
		return nil, false
	}
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		c.Delete(key)
		return nil, false
	}
	return item.value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// startEviction runs a goroutine to periodically clean up expired items
func (c *Cache) startEviction() {
	for {
		time.Sleep(1 * time.Minute)
		now := time.Now().UnixNano()
		c.mu.Lock()
		for key, item := range c.items {
			if item.expiration > 0 && now > item.expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func main() {
	cache := NewCache()
	cache.Set("key1", "value1", 5*time.Second)
	cache.Set("key2", "value2", 0) // no expiration

	value, found := cache.Get("key1")
	if found {
		fmt.Println("key1:", value)
	} else {
		fmt.Println("key1 not found")
	}

	time.Sleep(6 * time.Second)
	value, found = cache.Get("key1")
	if found {
		fmt.Println("key1:", value)
	} else {
		fmt.Println("key1 not found after expiration")
	}

	value, found = cache.Get("key2")
	if found {
		fmt.Println("key2:", value)
	} else {
		fmt.Println("key2 not found")
	}
}
