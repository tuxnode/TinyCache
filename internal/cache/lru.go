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
	c.lru.AddElement(key, value)
}

func (c *Cache) Get(key string) (value []byte, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return nil, false
	}

	if ele, ok := c.lru.cache[key]; ok {
		c.lru.ll.MoveToFront(ele)
		c.lru.stats.Hits += 1
		return ele.Value.(*entry).value.([]byte), true
	}
	c.lru.stats.Misses += 1
	return
}

// 仅仅查看缓存内容，不切换命中顺序
func (c *Cache) Peek(key string) ([]byte, bool) {
	if ele, ok := c.lru.cache[key]; ok {
		kv := ele.Value.(*entry)
		return kv.value.([]byte), true
	}
	return nil, false
}

func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return false
	}
	return c.lru.DeleteElement(key)
}

// 返回有效条目数量
func (c *Cache) Len() int {
	if c.lru == nil {
		return 0
	}
	return c.lru.LenLru()
}

func (c *Cache) Clear() {
	if c.lru == nil {
		return
	}
	c.lru.ClearElement()
}

func (c *Cache) Stats() Stats {
	if c.lru == nil {
		return Stats{}
	}

	c.lru.stats.Bytes = c.lru.nbyte
	c.lru.stats.Entries = c.lru.ll.Len()
	return c.lru.stats
}
