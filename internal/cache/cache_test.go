package cache

import (
	"bytes"
	"reflect"
	"testing"
)

func TestCacheZeroValueStartsEmpty(t *testing.T) {
	var c Cache

	t.Log("case: zero-value cache should behave like an empty cache")

	if got := callLen(t, &c); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}

	if value, ok := c.Get("missing"); ok {
		t.Fatalf("Get(missing) ok = true, value = %q, want miss", value)
	}

	if deleted := callDelete(t, &c, "missing"); deleted {
		t.Fatal("Delete(missing) = true, want false")
	}
}

func TestCacheAddGetAndUpdateExistingKey(t *testing.T) {
	var c Cache

	t.Log("case: add one key and read it back")
	c.Add("name", []byte("test"))

	value, ok := c.Get("name")
	if !ok {
		t.Fatal("Get(name) = miss, want hit")
	}
	if !bytes.Equal(value, []byte("test")) {
		t.Fatalf("Get(name) = %q, want %q", value, []byte("test"))
	}

	t.Log("case: update the same key and keep entry count unchanged")
	c.Add("name", []byte("next"))

	value, ok = c.Get("name")
	if !ok {
		t.Fatal("Get(name) after update = miss, want hit")
	}
	if !bytes.Equal(value, []byte("next")) {
		t.Fatalf("Get(name) after update = %q, want %q", value, []byte("next"))
	}

	if got := callLen(t, &c); got != 1 {
		t.Fatalf("Len() after updating the same key = %d, want 1", got)
	}

	stats := callStats(t, &c)
	t.Logf("stats after update: %+v", stats)
	if stats["Entries"] != 1 {
		t.Fatalf("Stats().Entries = %d, want 1", stats["Entries"])
	}
	if stats["Bytes"] != len("name")+len("next") {
		t.Fatalf("Stats().Bytes = %d, want %d", stats["Bytes"], len("name")+len("next"))
	}
}

func TestGetRefreshesRecency(t *testing.T) {
	c := &Cache{cacheBytes: 4}

	t.Log("case: Get should refresh recency before capacity eviction")
	c.Add("a", []byte("1"))
	c.Add("b", []byte("2"))

	if _, ok := c.Get("a"); !ok {
		t.Fatal("Get(a) = miss, want hit")
	}

	c.Add("c", []byte("3"))

	if _, ok := c.Get("b"); ok {
		t.Fatal("Get(b) = hit, want eviction after c is added")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("Get(a) = miss, want a to stay because Get should refresh recency")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("Get(c) = miss, want hit")
	}
}

func TestPeekDoesNotRefreshRecency(t *testing.T) {
	c := &Cache{cacheBytes: 4}

	t.Log("case: Peek should not refresh recency")
	c.Add("a", []byte("1"))
	c.Add("b", []byte("2"))

	value, ok := callPeek(t, c, "a")
	if !ok {
		t.Fatal("Peek(a) = miss, want hit")
	}
	if !bytes.Equal(value, []byte("1")) {
		t.Fatalf("Peek(a) = %q, want %q", value, []byte("1"))
	}

	c.Add("c", []byte("3"))

	if _, ok := c.Get("a"); ok {
		t.Fatal("Get(a) = hit, want a to be evicted because Peek must not refresh recency")
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("Get(b) = miss, want b to remain")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("Get(c) = miss, want c to remain")
	}
}

func TestDeleteRemovesEntryAndReportsPresence(t *testing.T) {
	var c Cache

	t.Log("case: Delete should remove an existing key and report existence correctly")
	c.Add("a", []byte("1"))
	c.Add("b", []byte("2"))

	if deleted := callDelete(t, &c, "a"); !deleted {
		t.Fatal("Delete(a) = false, want true")
	}

	if _, ok := c.Get("a"); ok {
		t.Fatal("Get(a) = hit, want miss after Delete")
	}

	if deleted := callDelete(t, &c, "a"); deleted {
		t.Fatal("Delete(a) the second time = true, want false")
	}

	if got := callLen(t, &c); got != 1 {
		t.Fatalf("Len() after Delete = %d, want 1", got)
	}
}

func TestClearRemovesAllEntries(t *testing.T) {
	var c Cache

	t.Log("case: Clear should remove all entries and reset live byte usage")
	c.Add("a", []byte("1"))
	c.Add("b", []byte("22"))

	if got := callLen(t, &c); got != 2 {
		t.Fatalf("Len() before Clear = %d, want 2", got)
	}

	callClear(t, &c)

	if got := callLen(t, &c); got != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", got)
	}

	if _, ok := c.Get("a"); ok {
		t.Fatal("Get(a) = hit, want miss after Clear")
	}
	if _, ok := c.Get("b"); ok {
		t.Fatal("Get(b) = hit, want miss after Clear")
	}

	stats := callStats(t, &c)
	t.Logf("stats after clear: %+v", stats)
	if stats["Entries"] != 0 {
		t.Fatalf("Stats().Entries after Clear = %d, want 0", stats["Entries"])
	}
	if stats["Bytes"] != 0 {
		t.Fatalf("Stats().Bytes after Clear = %d, want 0", stats["Bytes"])
	}
}

func TestStatsTracksCacheActivity(t *testing.T) {
	c := &Cache{cacheBytes: 4}

	t.Log("case: Stats should track hits, misses, evictions, deletes, entries, and bytes")
	c.Add("a", []byte("1"))
	c.Add("b", []byte("2"))

	if _, ok := c.Get("a"); !ok {
		t.Fatal("Get(a) = miss, want hit")
	}
	if _, ok := c.Get("z"); ok {
		t.Fatal("Get(z) = hit, want miss")
	}

	c.Add("c", []byte("3"))

	if deleted := callDelete(t, c, "a"); !deleted {
		t.Fatal("Delete(a) = false, want true")
	}

	stats := callStats(t, c)
	t.Logf("stats snapshot: %+v", stats)
	if stats["Hits"] != 1 {
		t.Fatalf("Stats().Hits = %d, want 1", stats["Hits"])
	}
	if stats["Misses"] != 1 {
		t.Fatalf("Stats().Misses = %d, want 1", stats["Misses"])
	}
	if stats["Evictions"] != 1 {
		t.Fatalf("Stats().Evictions = %d, want 1", stats["Evictions"])
	}
	if stats["Deletes"] != 1 {
		t.Fatalf("Stats().Deletes = %d, want 1", stats["Deletes"])
	}
	if stats["Entries"] != 1 {
		t.Fatalf("Stats().Entries = %d, want 1", stats["Entries"])
	}
	if stats["Bytes"] != 2 {
		t.Fatalf("Stats().Bytes = %d, want 2", stats["Bytes"])
	}
}

func callLen(t *testing.T, c *Cache) int {
	t.Helper()

	results := callMethod(t, c, "Len")
	if len(results) != 1 {
		t.Fatalf("Len() returned %d values, want 1", len(results))
	}

	return asInt(t, results[0], "Len()")
}

func callDelete(t *testing.T, c *Cache, key string) bool {
	t.Helper()

	results := callMethod(t, c, "Delete", key)
	if len(results) != 1 {
		t.Fatalf("Delete() returned %d values, want 1", len(results))
	}
	if results[0].Kind() != reflect.Bool {
		t.Fatalf("Delete() returned %s, want bool", results[0].Kind())
	}

	return results[0].Bool()
}

func callPeek(t *testing.T, c *Cache, key string) ([]byte, bool) {
	t.Helper()

	results := callMethod(t, c, "Peek", key)
	if len(results) != 2 {
		t.Fatalf("Peek() returned %d values, want 2", len(results))
	}

	value, ok := results[0].Interface().([]byte)
	if !ok {
		t.Fatalf("Peek() first return type = %T, want []byte", results[0].Interface())
	}
	if results[1].Kind() != reflect.Bool {
		t.Fatalf("Peek() second return kind = %s, want bool", results[1].Kind())
	}

	return value, results[1].Bool()
}

func callClear(t *testing.T, c *Cache) {
	t.Helper()

	results := callMethod(t, c, "Clear")
	if len(results) != 0 {
		t.Fatalf("Clear() returned %d values, want 0", len(results))
	}
}

func callStats(t *testing.T, c *Cache) map[string]int {
	t.Helper()

	results := callMethod(t, c, "Stats")
	if len(results) != 1 {
		t.Fatalf("Stats() returned %d values, want 1", len(results))
	}

	statsValue := results[0]
	if statsValue.Kind() == reflect.Pointer {
		if statsValue.IsNil() {
			t.Fatal("Stats() returned a nil pointer")
		}
		statsValue = statsValue.Elem()
	}
	if statsValue.Kind() != reflect.Struct {
		t.Fatalf("Stats() returned %s, want struct", statsValue.Kind())
	}

	fields := []string{"Hits", "Misses", "Evictions", "Deletes", "Entries", "Bytes"}
	stats := make(map[string]int, len(fields))
	for _, field := range fields {
		value := statsValue.FieldByName(field)
		if !value.IsValid() {
			t.Fatalf("Stats() result is missing field %s", field)
		}
		stats[field] = asInt(t, value, "Stats()."+field)
	}

	return stats
}

func callMethod(t *testing.T, recv any, name string, args ...any) []reflect.Value {
	t.Helper()

	method := reflect.ValueOf(recv).MethodByName(name)
	if !method.IsValid() {
		t.Fatalf("%T is missing method %s", recv, name)
	}

	if method.Type().NumIn() != len(args) {
		t.Fatalf("%s expects %d args, want %d", name, method.Type().NumIn(), len(args))
	}

	inputs := make([]reflect.Value, len(args))
	for i, arg := range args {
		inputs[i] = reflect.ValueOf(arg)
	}

	return method.Call(inputs)
}

func asInt(t *testing.T, value reflect.Value, label string) int {
	t.Helper()

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int(value.Uint())
	default:
		t.Fatalf("%s has kind %s, want an integer kind", label, value.Kind())
		return 0
	}
}
