package napping

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

var starter sync.Once
var addr net.Addr

func testHandler(w http.ResponseWriter, req *http.Request) {
	//make 500ms latency
	time.Sleep(500 * time.Millisecond)
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
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen - %s", err.Error())
	}
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
	//set 200ms connection timeout and 1000ms request timeout
	httpClient := NewClientSupportTimeout(200*time.Millisecond, 1000*time.Millisecond)
	req, _ := http.NewRequest("GET", "http://"+addr.String()+"/normal", nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		t.Fatalf("1st request should be successful")
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(data))

	req, _ = http.NewRequest("GET", "http://"+addr.String()+"/timeout", nil)
	resp, err = httpClient.Do(req)
	if err != nil {
		fmt.Println("is err expected?", err.Error())
	} else if err == nil {
		t.Fatalf("request should be timed out")
	}
}
