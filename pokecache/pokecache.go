package pokecache

import (
	"time"
	"sync"
)

type Cache struct {
	entries map[string]cacheEntry // name the map
	mutex sync.Mutex 			  // This is for synchronization
	interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time // This is the time the cache entry was created
	val []byte // this is the raw value of the cahce entry
}

func NewCache (interval time.Duration) *Cache {
	cache := Cache{
		entries: make(map[string]cacheEntry),
		interval: interval
	}
	go cache.reapLoop()
	return &cache
}

// Added method receivers - the (c *Cache) is the method receiver

func (c *Cache) Add(key string, val []byte) {
	// This needs to add the requested data to the cache
	// It should be referenced somewhere after the first response
	// Then store the data where ever... Probs in a variable/map

	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := cacheEntry{
		createdAt: time.Now(),
		val: val,
	}
	c.entries[key] = entry
	


}

func (c *Cache) Get(key string) ([]byte, bool) {
	// This should be run to check if the data requested in in the variable/cachce
	// This should be done somewhere prior to the get request, then default to the get request if it isn't located
	c.mutex.Lock()
	defer c.mutex.Unlock()
	entry, exists := c.entries[key] 
	if exists {
		return entry.val, true
	}
	return nil, false
	

}

func (C *Cache) reapLoop() {
	// This should use a timer to delete the data from the cache if it hasn't been used in a little while
	ticker := time.NewTicker(c.interval)
	for {
		<- ticker.C

		c.mutex.Lock()
		for key, entry := range c.entries {
			if time.Since(entry.createdAt) > c.interval {
				delete(c.entries, key)
			}
		}
		c.mutex.Unlock()
		
	}
}