package proxy

import (
	"net"
	"time"
)

type TimeoutConn struct {
	net.Conn
	Timeout time.Duration
}

func (tc *TimeoutConn) Read(buf []byte) (n int, err error) {
	if tc.Timeout > 0 {
		t := time.Now().Add(tc.Timeout)
		if err = tc.SetReadDeadline(t); err != nil {
			return
		}
	}
	n, err = tc.Conn.Read(buf)
	return
}

func (tc *TimeoutConn) Write(buf []byte) (n int, err error) {
	if tc.Timeout > 0 {
		t := time.Now().Add(tc.Timeout)
		if err = tc.SetWriteDeadline(t); err != nil {
			return
		}
	}
	n, err = tc.Conn.Write(buf)
	return
}
