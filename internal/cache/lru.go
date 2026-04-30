package cache

import (
	"sync"
)

type Cache struct {
	mu         sync.Mutex
	lru        *lru
	cacheBytes int
}

func (c *Cache) Add(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = NewLUR(c.cacheBytes)
	}
	c.Add(key, value)
}

func (c *Cache) Get(key string) (value []byte, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}

	if ele, ok := c.lru.cache[key]; ok {
		c.lru.ll.MoveToFront(ele)
		return ele.Value.(*entry).value.([]byte), true
	}
	return
}
