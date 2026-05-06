# TinyCache 测试说明

当前仓库已经包含一套以行为规格为导向的测试，文件位于 `internal/cache/cache_test.go`。

这些测试不会给出实现，只会定义你的 LRU 缓存必须满足的行为。

## 目标 API

测试期望 `Cache` 上至少实现以下方法：

```go
func (c *Cache) Add(key string, value []byte)
func (c *Cache) Get(key string) ([]byte, bool)
func (c *Cache) Peek(key string) ([]byte, bool)
func (c *Cache) Delete(key string) bool
func (c *Cache) Len() int
func (c *Cache) Clear()
func (c *Cache) Stats() Stats
```

`Stats` 的返回值应该是一个结构体，并至少包含这些整数字段：

```go
type Stats struct {
	Hits      int
	Misses    int
	Evictions int
	Deletes   int
	Entries   int
	Bytes     int
}
```

## 行为约定

1. `Cache{}` 的零值必须可以直接使用。
2. `Add` 需要能写入数据；当重复写入同一个 key 时，条目数不能增加。
3. `Get` 需要返回已存储的值，并刷新该元素的最近使用顺序。
4. `Peek` 需要返回已存储的值，但不能刷新最近使用顺序。
5. `Delete` 需要删除指定 key，并且只有在 key 原本存在时才返回 `true`。
6. `Len` 返回的是当前有效条目数，不是字节数。
7. `Clear` 需要清空所有有效条目，并把当前已使用字节数重置为 0。
8. 当缓存超过 `cacheBytes` 限制时，LRU 必须淘汰最久未使用的条目。
9. 单个条目的字节大小按 `len(key) + len(value)` 计算。
10. `Stats` 需要记录命中数、未命中数、淘汰数、主动删除数、当前条目数、当前字节数。

## 测试说明

1. 测试中会直接使用 `&Cache{cacheBytes: n}` 构造带容量限制的缓存。
2. 测试通过反射调用新增方法，这样在方法尚未实现时，测试会报出明确的失败信息，而不是直接编译错误。
3. `Delete` 只有在真正删除成功时才应该增加 `Deletes` 计数；删除失败时不应增加。
4. `Stats` 中的 `Entries` 和 `Bytes` 表示当前活跃数据，不是历史累计值。

## 运行方式

执行：

```bash
go test ./...
```

## 当前编译阻塞点

`internal/cache/lru.go` 里目前有一个未完成的 `LenLru` 方法，它没有返回值。

在运行测试之前，你需要先补完它，或者移除这个编译阻塞点。
