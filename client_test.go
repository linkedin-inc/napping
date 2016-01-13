package napping

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

var starter sync.Once
var addr net.Addr

func testHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func testDelayedHandler(w http.ResponseWriter, req *http.Request) {
	//make 2000ms latency
	time.Sleep(2000 * time.Millisecond)
	io.WriteString(w, "hello, world!\n")
}

func setupMockServer(t *testing.T) {
	http.HandleFunc("/normal", testHandler)
	http.HandleFunc("/timeout", testDelayedHandler)
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf("failed to listen - %s", err.Error())
	}
	fmt.Sprintf("listened at - %s", ln.Addr())
	go func() {
		err = http.Serve(ln, nil)
		if err != nil {
			t.Fatalf("failed to start HTTP server - %s", err.Error())
		}
	}()
	addr = ln.Addr()
}

func TestNewClientSupportTimeout(t *testing.T) {
	starter.Do(func() { setupMockServer(t) })
	fmt.Println("listen at - ", addr.String())
	httpClient := NewClientSupportTimeout(500 * time.Millisecond)
	req, _ := http.NewRequest("GET", "http://"+addr.String()+"/normal", nil)
	resp, err := httpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("1st request should be successful")
	}
	var channel chan int
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				req, _ = http.NewRequest("GET", "http://"+addr.String()+"/timeout", nil)
				begin := time.Now()
				resp, err = httpClient.Do(req)
				end := time.Now()
				if err != nil {
					fmt.Println("is err expected?", err.Error(), "cost:", end.Sub(begin).Nanoseconds())
					channel <- i
				} else if err == nil {
					t.Fatalf("request should be timed out")
				}
			}
		}()
	}
	<-channel
}
