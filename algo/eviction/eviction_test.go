package eviction

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLength(t *testing.T) {
	f := func(gc Eviction) {
		gc.Set("test1", 0)
		gc.Set("test2", 0)
		assert.Equalf(t, 2, gc.Len(true), "Expected length is %v, not 2", gc.Len(true))
	}
	t.Run("LFU", func(t *testing.T) { f(NewLFUCache(1000)) })
	t.Run("LRU", func(t *testing.T) { f(NewLRUCache(1000)) })
}

func TestLFUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11

	f := func(gc Eviction) {
		for i := 0; i < numbers; i++ {
			gc.Set(fmt.Sprintf("Key-%d", i), i)
			_, err := gc.Get(fmt.Sprintf("Key-%d", i))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	}
	t.Run("LFU", func(t *testing.T) { f(NewLFUCache(cacheSize)) })
	t.Run("LRU", func(t *testing.T) { f(NewLRUCache(cacheSize)) })
}

func TestLFUHas(t *testing.T) {
	f := func(gc Eviction) {
		gc.SetExpiration(10 * time.Millisecond)

		for i := 0; i < 10; i++ {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				gc.Set("test1", 0)
				gc.Set("test2", 0)
				gc.Get("test1")
				gc.Get("test2")

				if gc.Has("test0") {
					t.Fatal("should not have test0")
				}
				if !gc.Has("test1") {
					t.Fatal("should have test1")
				}
				if !gc.Has("test2") {
					t.Fatal("should have test2")
				}

				time.Sleep(20 * time.Millisecond)

				if gc.Has("test0") {
					t.Fatal("should not have test0")
				}
				if gc.Has("test1") {
					t.Fatal("should not have test1")
				}
				if gc.Has("test2") {
					t.Fatal("should not have test2")
				}
			})
		}
	}

	t.Run("LFU", func(t *testing.T) { f(NewLFUCache(2)) })
	t.Run("LRU", func(t *testing.T) { f(NewLRUCache(2)) })
}

func TestLFUFreqListOrder(t *testing.T) {
	gc := NewLFUCache(5)
	for i := 4; i >= 0; i-- {
		v := strconv.Itoa(i)
		gc.Set(v, i)
		for j := 0; j <= i; j++ {
			gc.Get(v)
		}
	}
	if l := gc.(*LFUCache).freqList.Len(); l != 6 {
		t.Fatalf("%v != 6", l)
	}
	var i uint
	for e := gc.(*LFUCache).freqList.Front(); e != nil; e = e.Next() {
		if e.Value.(*freqEntry).freq != i {
			t.Fatalf("%v != %v", e.Value.(*freqEntry).freq, i)
		}
		i++
	}
	gc.Remove("1")

	if l := gc.(*LFUCache).freqList.Len(); l != 5 {
		t.Fatalf("%v != 5", l)
	}
	gc.Set("1", 1)
	if l := gc.(*LFUCache).freqList.Len(); l != 5 {
		t.Fatalf("%v != 5", l)
	}
	gc.Get("1")
	if l := gc.(*LFUCache).freqList.Len(); l != 5 {
		t.Fatalf("%v != 5", l)
	}
	gc.Get("1")
	if l := gc.(*LFUCache).freqList.Len(); l != 6 {
		t.Fatalf("%v != 6", l)
	}
}

func TestLFUFreqListLength(t *testing.T) {
	k0, v0 := "k0", "v0"
	k1, v1 := "k1", "v1"

	{
		gc := NewLFUCache(5)
		if l := gc.(*LFUCache).freqList.Len(); l != 1 {
			t.Fatalf("%v != 1", l)
		}
	}
	{
		gc := NewLFUCache(5)
		gc.Set(k0, v0)
		for i := 0; i < 5; i++ {
			gc.Get(k0)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}

	{
		gc := NewLFUCache(5)
		gc.Set(k0, v0)
		gc.Set(k1, v1)
		for i := 0; i < 5; i++ {
			gc.Get(k0)
			gc.Get(k1)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}

	{
		gc := NewLFUCache(5)
		gc.Set(k0, v0)
		gc.Set(k1, v1)
		for i := 0; i < 5; i++ {
			gc.Get(k0)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
		for i := 0; i < 5; i++ {
			gc.Get(k1)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}

	{
		gc := NewLFUCache(5)
		gc.Set(k0, v0)
		gc.Get(k0)
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
		gc.Remove(k0)
		if l := gc.(*LFUCache).freqList.Len(); l != 1 {
			t.Fatalf("%v != 1", l)
		}
		gc.Set(k0, v0)
		if l := gc.(*LFUCache).freqList.Len(); l != 1 {
			t.Fatalf("%v != 1", l)
		}
		gc.Get(k0)
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}
}
