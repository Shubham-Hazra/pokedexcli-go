package pokecache

import (
	"log"
	"sync"
	"time"
)

type CacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entries map[string]CacheEntry
	mu      *sync.RWMutex
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[string]CacheEntry)
	}

	if _, ok := c.entries[key]; ok {
		log.Printf("Key: %v, already exists in cache", key)
		return
	}

	entry := CacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

	c.entries[key] = entry
	log.Printf("Added key: %v to cache", key)
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.entries[key]
	if !ok {
		log.Printf("Key: %v not found in cache", key)
		return nil, false
	}

	log.Printf("Retrieved key: %v from cache", key)
	return val.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for k, v := range c.entries {
			if time.Since(v.createdAt) > interval {
				delete(c.entries, k)
				log.Printf("Removed expired key: %v from cache", k)
			}
		}
		c.mu.Unlock()
	}
}

func NewCache(interval time.Duration) Cache {
	c := Cache{}
	if c.entries == nil {
		c.entries = make(map[string]CacheEntry)
	}
	c.mu = &sync.RWMutex{}

	log.Printf("Cache initialized with reaping interval: %v", interval)

	go c.reapLoop(interval)
	return c
}
