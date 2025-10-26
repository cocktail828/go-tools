package gcache

import (
	"fmt"
	"testing"
	"time"
)

func loader(key any) (any, error) {
	return fmt.Sprintf("valueFor%s", key), nil
}

func testSetCache(t *testing.T, gc Cache, numbers int) {
	for i := 0; i < numbers; i++ {
		key := fmt.Sprintf("Key-%d", i)
		value, err := loader(key)
		if err != nil {
			t.Error(err)
			return
		}
		gc.Set(key, value)
	}
}

func testGetCache(t *testing.T, gc Cache, numbers int) {
	for i := 0; i < numbers; i++ {
		key := fmt.Sprintf("Key-%d", i)
		v, err := gc.Get(key)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expectedV, _ := loader(key)
		if v != expectedV {
			t.Errorf("Expected value is %v, not %v", expectedV, v)
		}
	}
}

func setItemsByRange(t *testing.T, c Cache, start, end int) {
	for i := start; i < end; i++ {
		if err := c.Set(i, i); err != nil {
			t.Error(err)
		}
	}
}

func keysToMap(keys []any) map[any]struct{} {
	m := make(map[any]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return m
}

func checkItemsByRange(t *testing.T, keys []any, m map[any]any, size, start, end int) {
	if len(keys) != size {
		t.Fatalf("%v != %v", len(keys), size)
	} else if len(m) != size {
		t.Fatalf("%v != %v", len(m), size)
	}
	km := keysToMap(keys)
	for i := start; i < end; i++ {
		if _, ok := km[i]; !ok {
			t.Errorf("keys should contain %v", i)
		}
		v, ok := m[i]
		if !ok {
			t.Errorf("m should contain %v", i)
			continue
		}
		if v.(int) != i {
			t.Errorf("%v != %v", v, i)
			continue
		}
	}
}

func testExpiredItems(t *testing.T, evT EvictType) {
	size := 8
	cache := New(size).
		Expiration(time.Millisecond).
		EvictType(evT).
		Build()

	setItemsByRange(t, cache, 0, size)
	checkItemsByRange(t, cache.Keys(true), cache.GetALL(true), cache.Len(true), 0, size)

	time.Sleep(time.Millisecond)

	checkItemsByRange(t, cache.Keys(false), cache.GetALL(false), cache.Len(false), 0, size)

	if l := cache.Len(true); l != 0 {
		t.Fatalf("GetALL should returns no items, but got length %v", l)
	}

	cache.Set(1, 1)
	m := cache.GetALL(true)
	if len(m) != 1 {
		t.Fatalf("%v != %v", len(m), 1)
	} else if l := cache.Len(true); l != 1 {
		t.Fatalf("%v != %v", l, 1)
	}
	if m[1] != 1 {
		t.Fatalf("%v != %v", m[1], 1)
	}
}

func getSimpleEvictedFunc(t *testing.T) func(any, any) {
	return func(key, value any) {
		t.Logf("Key=%v Value=%v will be evicted.\n", key, value)
	}
}

func buildTestCache(t *testing.T, tp EvictType, size int) Cache {
	return New(size).
		EvictType(tp).
		EvictedFunc(getSimpleEvictedFunc(t)).
		Build()
}

func buildTestLoadingCache(t *testing.T, tp EvictType, size int, loader LoaderFunc) Cache {
	return New(size).
		EvictType(tp).
		LoaderFunc(loader).
		EvictedFunc(getSimpleEvictedFunc(t)).
		Build()
}

func buildTestLoadingCacheWithExpiration(t *testing.T, tp EvictType, size int, ep time.Duration) Cache {
	return New(size).
		EvictType(tp).
		Expiration(ep).
		LoaderFunc(loader).
		EvictedFunc(getSimpleEvictedFunc(t)).
		Build()
}
