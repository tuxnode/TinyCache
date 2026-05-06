package cache

import (
	"container/list"
)

type entry struct {
	key   string
	value any
}

type lru struct {
	maxByte int
	nbyte   int // 已写入字符
	ll      *list.List
	cache   map[string]*list.Element
	stats   Stats
}

type Stats struct {
	Hits      int // 缓存命中次数
	Misses    int // 未命中次数
	Evictions int // 被LRU主动淘汰的次数
	Deletes   int // 主动删除成功的次数
	Entries   int // 当前缓存里还活着的条目数
	Bytes     int // 当前缓存里还活着的总字节数
}

/* Init Lur */
func NewLUR(maxByte int) *lru {
	return &lru{
		maxByte: maxByte,
		ll:      list.New(),
		cache:   make(map[string]*list.Element),
		stats:   Stats{},
	}
}

/* 向Lur缓存添加数据 */
func (c *lru) AddElement(key string, value any) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		// 类型断言
		kv := ele.Value.(*entry)
		c.nbyte += len(value.([]byte)) - len(kv.value.([]byte))
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbyte += len(key) + len(value.([]byte))
	}

	for c.maxByte != 0 && c.maxByte < c.nbyte {
		c.RemoveOldest()
	}
}

func (c *lru) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbyte -= len(kv.key) + len(kv.value.([]byte))

		// 触发超量淘汰
		c.stats.Evictions += 1
	}
}

func (c *lru) DeleteElement(key string) bool {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		delete(c.cache, key)

		kv := ele.Value.(*entry)
		c.nbyte -= len(kv.key) + len(kv.value.([]byte))

		c.stats.Deletes += 1
		return true
	}
	return false
}

func (c *lru) LenLru() int {
	return c.ll.Len()
}

func (c *lru) ClearElement() {
	c.ll.Init()
	c.cache = make(map[string]*list.Element)
	c.nbyte = 0
	c.stats = Stats{}
}
