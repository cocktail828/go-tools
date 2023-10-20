package pool_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/pool"
	"github.com/cocktail828/go-tools/algo/pool/driver"
	"github.com/cocktail828/go-tools/z"
	"github.com/pkg/errors"
)

type Conn struct {
	conn net.Conn
}

func (c *Conn) Ping(ctx context.Context) error {
	fmt.Println("conn Ping")
	return nil
}

func (c *Conn) Close() error {
	fmt.Println("conn Close")
	return c.conn.Close()
}

func (c *Conn) ResetSession(ctx context.Context) error {
	fmt.Println("conn ResetSession")
	return nil
}

func (c *Conn) IsValid() bool {
	fmt.Println("conn IsValid")
	return true
}

type FakeDriver struct{}

func (d *FakeDriver) Open(ctx context.Context, name string) (driver.Conn, error) {
	fmt.Println("driver Open", name)
	conn, err := net.Dial("tcp", name)
	return &Conn{conn}, err
}

func init() {
	pool.Register("fake", &FakeDriver{})
}

func TestPool(t *testing.T) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	defer db.Close()

	z.Must(db.Ping())
	fmt.Printf("%#v\n", db.Stats())

	for i := 0; i < 10; i++ {
		z.Must(db.DoContext(context.Background(), func(ctx context.Context, ci driver.Conn) error {
			if ci == nil {
				return errors.Errorf("unknow ci")
			}
			return nil
		}))
	}
	fmt.Printf("%#v\n", db.Stats())
}

func TestPoolParam(t *testing.T) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	defer db.Close()

	// z.Must(db.Ping())
	// fmt.Printf("%#v\n", db.Stats())
	db.SetConnMaxIdleTime(time.Second)
	for i := 0; i < 3; i++ {
		z.Must(db.DoContext(context.Background(), func(ctx context.Context, ci driver.Conn) error {
			if ci == nil {
				return errors.Errorf("unknow ci")
			}
			return nil
		}))
		time.Sleep(time.Millisecond * 1500)
	}
	fmt.Printf("%#v\n", db.Stats())
}

func BenchmarkPool(b *testing.B) {
	db, err := pool.Open("fake", "10.1.87.70:1337")
	z.Must(err)
	defer db.Close()

	fmt.Printf("%#v\n", db.Stats())
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			z.Must(db.DoContext(context.Background(), func(ctx context.Context, ci driver.Conn) error {
				time.Sleep(time.Millisecond * 3000)
				if ci == nil {
					panic("unknow ci")
				}
				return nil
			}))
		}
	})
	fmt.Printf("%#v\n", db.Stats())
}
