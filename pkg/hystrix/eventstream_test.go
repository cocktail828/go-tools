package hystrix

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// )

// type eventStreamTestServer struct {
// 	*httptest.Server
// 	*StreamHandler
// }

// func (s *eventStreamTestServer) stopTestServer() error {
// 	s.Close()
// 	s.Stop()
// 	Flush()

// 	return nil
// }

// func startTestServer() *eventStreamTestServer {
// 	hystrixStreamHandler := NewStreamHandler()
// 	hystrixStreamHandler.Start()
// 	return &eventStreamTestServer{
// 		httptest.NewServer(hystrixStreamHandler),
// 		hystrixStreamHandler,
// 	}
// }

// func sleepingCommand(t *testing.T, name string, duration time.Duration) {
// 	done := make(chan bool)
// 	errChan := Go(name, func() error {
// 		time.Sleep(duration)
// 		done <- true
// 		return nil
// 	}, nil)

// 	select {
// 	case <-done:
// 		// do nothing
// 	case err := <-errChan:
// 		t.Fatal(err)
// 	}
// }

// func failingCommand(t *testing.T, name string, duration time.Duration) {
// 	done := make(chan bool)
// 	errChan := Go(name, func() error {
// 		time.Sleep(duration)
// 		return fmt.Errorf("fail")
// 	}, nil)

// 	select {
// 	case <-done:
// 		t.Fatal("should not have succeeded")
// 	case <-errChan:
// 		// do nothing
// 	}
// }

// // grabFirstFromStream reads on the http request until we see the first
// // full result printed
// func grabFirstCommandFromStream(t *testing.T, url string) streamCmdMetric {
// 	var event streamCmdMetric

// 	metrics, done := streamMetrics(t, url)
// 	for m := range metrics {
// 		if strings.Contains(m, "HystrixCommand") {
// 			done <- true
// 			close(done)

// 			err := json.Unmarshal([]byte(m), &event)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			break
// 		}
// 	}

// 	return event
// }

// func grabFirstThreadPoolFromStream(t *testing.T, url string) streamThreadPoolMetric {
// 	var event streamThreadPoolMetric

// 	metrics, done := streamMetrics(t, url)
// 	for m := range metrics {
// 		if strings.Contains(m, "HystrixThreadPool") {
// 			done <- true
// 			close(done)

// 			err := json.Unmarshal([]byte(m), &event)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			break
// 		}
// 	}

// 	return event
// }

// func streamMetrics(t *testing.T, url string) (chan string, chan bool) {
// 	metrics := make(chan string, 1)
// 	done := make(chan bool, 1)

// 	go func() {
// 		res, err := http.Get(url)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer res.Body.Close()

// 		buf := []byte{0}
// 		data := ""
// 		for {
// 			_, err := res.Body.Read(buf)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			data += string(buf)
// 			if strings.Contains(data, "\n\n") {
// 				data = strings.Replace(data, "data:{", "{", 1)
// 				metrics <- data
// 				data = ""
// 			}

// 			select {
// 			case <-done:
// 				close(metrics)
// 				return
// 			default:
// 			}
// 		}
// 	}()

// 	return metrics, done
// }

// func TestEventStream(t *testing.T) {
// 	server := startTestServer()
// 	defer server.stopTestServer()

// 	sleepingCommand(t, "eventstream", 1*time.Millisecond)
// 	sleepingCommand(t, "eventstream", 1*time.Millisecond)

// 	event := grabFirstCommandFromStream(t, server.URL)
// 	assert.Equal(t, "eventstream", event.Name)
// 	assert.Equal(t, 2, int(event.RequestCount))

// 	sleepingCommand(t, "errorpercent", 1*time.Millisecond)
// 	failingCommand(t, "errorpercent", 1*time.Millisecond)
// 	failingCommand(t, "errorpercent", 1*time.Millisecond)

// 	metric := grabFirstCommandFromStream(t, server.URL)
// 	assert.Equal(t, 67, metric.ErrorPct, "the error precentage should be 67")
// }

// func TestClientCancelEventStream(t *testing.T) {
// 	server := startTestServer()
// 	defer server.stopTestServer()

// 	sleepingCommand(t, "eventstream", 1*time.Millisecond)

// 	req, err := http.NewRequest("GET", server.URL, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// use a transport so we can cancel the stream when we're done - in 1.5 this is much easier
// 	tr := &http.Transport{}
// 	client := &http.Client{Transport: tr}
// 	wait := make(chan struct{})
// 	afterFirstRead := &sync.WaitGroup{}
// 	afterFirstRead.Add(1)

// 	go func() {
// 		afr := afterFirstRead
// 		buf := []byte{0}
// 		res, err := client.Do(req)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer res.Body.Close()

// 		for {
// 			select {
// 			case <-wait:
// 				//wait for master goroutine to break us out
// 				tr.CancelRequest(req)
// 				return
// 			default:
// 				//read something
// 				_, err = res.Body.Read(buf)
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 				if afr != nil {
// 					afr.Done()
// 					afr = nil
// 				}
// 			}
// 		}
// 	}()
// 	// need to make sure our request has round-tripped to the server
// 	afterFirstRead.Wait()

// 	server.StreamHandler.mu.RLock()
// 	assert.Equal(t, 1, len(server.StreamHandler.requests))
// 	server.StreamHandler.mu.RUnlock()
// 	// let the request be cancelled and the body closed
// 	close(wait)
// 	// wait for the server to clean up
// 	time.Sleep(2000 * time.Millisecond)
// 	//confirm we have 0 clients
// 	server.StreamHandler.mu.RLock()
// 	assert.Equal(t, 0, len(server.StreamHandler.requests), "it should be detected as disconnected and de-registered")
// 	server.StreamHandler.mu.RUnlock()
// }

// func TestThreadPoolStream(t *testing.T) {
// 	server := startTestServer()
// 	defer server.stopTestServer()

// 	sleepingCommand(t, "threadpool", 1*time.Millisecond)
// 	metric := grabFirstThreadPoolFromStream(t, server.URL)

// 	assert.Equal(t, 1, metric.RollingCountThreadsExecuted, "the rolling count of executions should increment")
// 	assert.Equal(t, 10, metric.CurrentPoolSize, "the pool size should be 10")
// }
