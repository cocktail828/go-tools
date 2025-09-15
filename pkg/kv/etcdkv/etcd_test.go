package etcdkv

import (
	"context"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/kv"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	endpoint = "172.29.231.108:12379"
	src      kv.KV
)

func TestMain(b *testing.M) {
	var err error
	src, err = New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: time.Second,
	}, "aaa/bbb")
	z.Must(err)
	b.Run()
}

type Case struct {
	IsPut bool
	Name  string
	Key   string
}

func (c Case) Event() kv.Event {
	ev := &etcdEvent{}
	if c.IsPut {
		ev.Append(kv.PUT, c.Key, []byte(c.Key))
	} else {
		ev.Append(kv.DELETE, c.Key, nil)
	}

	return ev
}

type Cases []Case

func (cs Cases) Convert() kv.Result {
	impl := &etcdKvPairs{}
	for _, c := range cs {
		impl.Append(c.Key, []byte(c.Key))
	}
	return impl
}

func TestEtcdCount(t *testing.T) {
	c := Case{Key: "a"}

	assert.NoError(t, src.Set(context.TODO(), c.Key, []byte(c.Key)))
	kv, err := src.Get(context.TODO(), c.Key, WithCount(), WithMatchPrefix())
	assert.NoError(t, err)
	assert.EqualValues(t, CountResult{Num: 1}, kv)
}

func TestEtcdTTL(t *testing.T) {
	c := Case{Key: "a"}

	assert.NoError(t, src.Set(context.TODO(), c.Key, []byte(c.Key), WithTTL(1)))
	kv, err := src.Get(context.TODO(), c.Key)
	assert.NoError(t, err)
	assert.EqualValues(t, &etcdKvPairs{Keys: []string{c.Key}, Values: [][]byte{[]byte(c.Key)}}, kv)

	// make sure key is expired
	time.Sleep(time.Millisecond * 3000)
	kv, err = src.Get(context.TODO(), c.Key)
	assert.NoError(t, err)
	assert.EqualValues(t, &etcdKvPairs{}, kv)
}

func TestEtcdKV(t *testing.T) {
	cases := Cases{
		{false, "kv", "a"},
		{false, "kv-tree", "a/123"},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			assert.NoError(t, src.Set(context.TODO(), c.Key, []byte(c.Key)))
			kv, err := src.Get(context.TODO(), c.Key)
			assert.NoError(t, err)
			assert.EqualValues(t, &etcdKvPairs{Keys: []string{c.Key}, Values: [][]byte{[]byte(c.Key)}}, kv)
		})
	}

	kv, err := src.Get(context.TODO(), "a", WithMatchPrefix())
	assert.NoError(t, err)
	assert.EqualValues(t, cases.Convert(), kv)
	assert.NoError(t, src.Del(context.TODO(), "a", DelWithPrefix()))
}

func TestEtcdWatchPrefix(t *testing.T) {
	cases := Cases{
		{true, "", "a"}, // PUT first
		{false, "", "a"},
		{true, "", "a/b"},
		{false, "", "a/b"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := src.Watch(ctx, WatchWithPrefix())
	for _, c := range cases {
		if c.IsPut {
			assert.NoError(t, src.Set(context.TODO(), c.Key, []byte(c.Key)))
		} else {
			assert.NoError(t, src.Del(context.TODO(), c.Key))
		}
	}

	for _, c := range cases {
		events, err := w.Next(context.Background())
		assert.NoError(t, err)
		assert.EqualValues(t, c.Event(), events)
	}
}
