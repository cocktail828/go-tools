package pool_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/pool"
	"github.com/cocktail828/go-tools/algo/pool/driver"
	"github.com/cocktail828/go-tools/z"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	gOpenCount = atomic.Int64{}
)

type Conn struct{ isOpen bool }

func (c *Conn) Ping(ctx context.Context) error {
	fmt.Println("conn Ping")
	return nil
}

func (c *Conn) Close() error {
	fmt.Println("conn Close")
	c.isOpen = false
	gOpenCount.Add(-1)
	return nil
}

func (c *Conn) ResetSession(ctx context.Context) error {
	// fmt.Println("conn ResetSession")
	return nil
}

func (c *Conn) IsValid() bool {
	// fmt.Println("conn IsValid")
	return true
}

type FakeDriver struct{}

func (d *FakeDriver) Open(ctx context.Context, name string) (driver.Conn, error) {
	fmt.Println("driver Open", name)
	gOpenCount.Add(1)
	return &Conn{isOpen: true}, nil
}

func init() {
	pool.Register("fake", &FakeDriver{})
}

func TestPool(t *testing.T) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	defer db.Close()
	z.Must(db.Ping())

	for i := 0; i < 3; i++ {
		z.Must(db.DoContext(context.Background(), func(ci driver.Conn) error {
			fmt.Printf("inner %#v\n", db.Stats())
			return nil
		}))
		fmt.Printf("outer %#v\n", db.Stats())
	}
}

func TestPoolDeadline(t *testing.T) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	db.CloseOnDeadline()
	defer db.Close()
	z.Must(db.Ping())

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	var c *Conn
	assert.Equal(t, context.DeadlineExceeded, db.DoContext(ctx, func(ci driver.Conn) error {
		if ci == nil {
			return errors.Errorf("unknow ci")
		}
		c = ci.(*Conn)
		time.Sleep(time.Hour)
		return nil
	}))
	assert.Equal(t, false, c.isOpen)
	assert.Equal(t, 0, db.Stats().OpenCount)
	assert.Equal(t, 0, db.Stats().IdleCount)
}

func TestPoolMaxConn(t *testing.T) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	defer db.Close()

	db.SetMaxOpenConns(3)
	wg := sync.WaitGroup{}
	wg.Add(3)
	wgc := sync.WaitGroup{}
	wgc.Add(3)

	for i := 0; i < 3; i++ {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		go func() {
			defer wgc.Done()
			assert.Equal(t, nil, db.DoContext(ctx, func(ci driver.Conn) error {
				wg.Done()
				time.Sleep(time.Second * 2)
				return nil
			}))
		}()
	}
	wg.Wait()

	fmt.Printf("%#v\n", db.Stats())
	assert.Equal(t, nil, db.Do(func(ci driver.Conn) error {
		return nil
	}))
	wgc.Wait()
	fmt.Printf("%#v\n", db.Stats())
}

func BenchmarkPool(b *testing.B) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	defer db.Close()

	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(time.Second)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			z.Must(db.DoContext(context.Background(), func(ci driver.Conn) error {
				if ci == nil {
					panic("unknow ci")
				}
				return nil
			}))
		}
	})
}
