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
}

/* Init Lur */
func NewLUR(maxByte int) *lru {
	return &lru{
		maxByte: maxByte,
		ll:      list.New(),
		cache:   make(map[string]*list.Element),
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
	}
}
