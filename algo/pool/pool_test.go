package pool_test

import (
	"context"
	"fmt"
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

type Conn struct{}

func (c *Conn) Ping(ctx context.Context) error {
	fmt.Println("conn Ping")
	return nil
}

func (c *Conn) Close() error {
	fmt.Println("conn Close")
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
	// fmt.Println("driver Open", name)
	gOpenCount.Add(1)
	return &Conn{}, nil
}

func init() {
	pool.Register("fake", &FakeDriver{})
}

func TestPool(t *testing.T) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	defer db.Close()
	z.Must(db.Ping())

	for i := 0; i < 1; i++ {
		z.Must(db.DoContext(context.Background(), func(ci driver.Conn) error {
			fmt.Printf("inner %#v\n", db.Stats())
			if ci == nil {
				return errors.Errorf("unknow ci")
			}
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
	assert.Equal(t, context.DeadlineExceeded, db.DoContext(ctx, func(ci driver.Conn) error {
		if ci == nil {
			return errors.Errorf("unknow ci")
		}
		time.Sleep(time.Hour)
		return nil
	}))
	assert.Equal(t, 0, db.Stats().OpenCount)
	assert.Equal(t, 0, db.Stats().IdleCount)
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
