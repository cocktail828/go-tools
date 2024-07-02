package etcd_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/kvstore"
	"github.com/cocktail828/go-tools/pkg/kvstore/etcd"
	"github.com/stretchr/testify/assert"
)

var (
	endpoint = "127.0.0.1:2379"
)

func TestEtcd(t *testing.T) {
	src, err := etcd.NewSource(etcd.WithAddress(endpoint),
		etcd.WithPrefix("/caesar/"),
		etcd.WithDialTimeout(time.Second*3))
	assert.Nil(t, err)
	defer src.Close()

	cases := []struct {
		Name string
		Key  string
		Val  []byte
	}{
		{"kv", "a", []byte("aaa")},
		{"kv-tree", "a/123", []byte("a/123")},
	}

	for _, c := range cases {
		assert.Equal(t, nil, src.Write(context.Background(), c.Key, c.Val))
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			kv, err := src.Read(context.Background(), c.Key)
			assert.Nil(t, err)
			assert.ElementsMatch(t, []kvstore.KVPair{{Key: c.Key, Val: c.Val}}, kv)
		})
	}

	kv, err := src.Read(context.Background(), "a", etcd.MatchPrefix())
	assert.Nil(t, err)
	assert.Equal(t, []kvstore.KVPair{{Key: "a", Val: []byte("aaa")}, {Key: "a/123", Val: []byte("a/123")}}, kv)
	assert.Equal(t, nil, src.Delete(context.Background(), "a", etcd.MatchPrefix()))
}

func TestEtcdWatch(t *testing.T) {
	cases := []struct {
		Name string
		Key  string
		Val  []byte
		Opts []kvstore.Option
	}{
		{"watchkv", "a", []byte("aaa"), nil},
		{"watchkv-del", "a", []byte(""), nil},
		{"watchkv-tree", "a/123", []byte("a/123"), []kvstore.Option{etcd.MatchPrefix()}},
		{"watchkv-tree-del", "a/123", []byte(""), []kvstore.Option{etcd.MatchPrefix()}},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			src, err := etcd.NewSource(etcd.WithAddress(endpoint),
				etcd.WithPrefix("/caesar/"),
				etcd.WithDialTimeout(time.Second*3))
			assert.Nil(t, err)
			defer src.Close()

			w := src.Watch(context.Background(), c.Key, c.Opts...)
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				defer cancel()
				for {
					events, err := w.Next()
					if err == io.EOF {
						return
					}
					if len(events) == 0 {
						continue
					}
					assert.Nil(t, err)
					if len(c.Val) == 0 {
						assert.Equal(t, []kvstore.Event{{Type: kvstore.Del, Key: c.Key}}, events)
					} else {
						assert.Equal(t, []kvstore.Event{{Type: kvstore.Put, Key: c.Key, Val: c.Val}}, events)
					}
				}
			}()

			for i := 0; i < 2; i++ {
				<-time.After(time.Millisecond * 100)
				if len(c.Val) == 0 {
					assert.Equal(t, nil, src.Delete(context.Background(), c.Key))
				} else {
					assert.Equal(t, nil, src.Write(context.Background(), c.Key, c.Val))
				}
			}
			w.Stop()
			<-ctx.Done()
		})
	}
}
