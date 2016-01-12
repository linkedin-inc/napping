package napping

import (
	"net"
	"net/http"
	"time"
)

const (
	DefaultConnTimeout = 200 * time.Millisecond
	DefaultReqTimeout  = 1000 * time.Millisecond
)

func newDialerSupportTimeout(connTimeout time.Duration, reqTimeout time.Duration) func(network, address string) (net.Conn, error) {
	return func(network, address string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, address, connTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(reqTimeout))
		return conn, nil
	}
}

func newTransportSupportTimeout(connTimeout time.Duration, reqTimeout time.Duration) *http.Transport {
	return &http.Transport{
		Dial: newDialerSupportTimeout(connTimeout, reqTimeout),
	}
}

//NewClientSupportTimeout return a http client support timeout, which you can specify connection timeout and request timeout.
func NewClientSupportTimeout(args ...interface{}) *http.Client {
	connTimeout := DefaultConnTimeout
	reqTimeout := DefaultReqTimeout
	if len(args) == 1 {
		connTimeout = args[0].(time.Duration)
	}
	if len(args) == 2 {
		connTimeout = args[0].(time.Duration)
		reqTimeout = args[1].(time.Duration)
	}
	return &http.Client{
		Transport: newTransportSupportTimeout(connTimeout, reqTimeout),
		Timeout:   DefaultConnTimeout + DefaultReqTimeout,
	}
}
