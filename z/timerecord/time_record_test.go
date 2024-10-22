package timerecord_test

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/timerecord"
)

func TestTimeRecorder(t *testing.T) {
	tr := timerecord.NewTimeRecorder()
	time.Sleep(time.Second)
	log.Println(tr.Duration(), tr.Elapse())

	time.Sleep(time.Second)
	log.Println(tr.Duration(), tr.Elapse())
}

func TestTimeOverseer(t *testing.T) {
	to := timerecord.NewTimeOverseer(time.Second)
	c := to.Start()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			<-time.After(time.Millisecond * 100)
			to.Reset()
		}
	}()

	go func() {
		defer wg.Done()
		for {
			_, ok := <-c
			if ok {
				log.Println("timer")
			} else {
				log.Println("quit")
				return
			}
		}
	}()
	time.Sleep(time.Second * 5)
	to.Stop()
	wg.Wait()
}
