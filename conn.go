package main

import (
	"net"
	"time"
)

type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

func newTimeoutConn(conn net.Conn, timeout time.Duration) net.Conn {
	tc := new(timeoutConn)
	tc.Conn = conn
	tc.timeout = timeout
	return tc
}

func (tc *timeoutConn) Read(buf []byte) (int, error) {
	d := time.Now().Add(tc.timeout)
	err := tc.Conn.SetDeadline(d)
	if err != nil {
		return 0, err
	}
	return tc.Conn.Read(buf)
}

func (tc *timeoutConn) Write(buf []byte) (int, error) {
	d := time.Now().Add(tc.timeout)
	err := tc.Conn.SetDeadline(d)
	if err != nil {
		return 0, err
	}
	return tc.Conn.Write(buf)
}

type cipherConn struct {
	net.Conn
	*Cipher
}

func newCipherConn(cipher *Cipher, conn net.Conn) (cc *cipherConn) {
	cc = new(cipherConn)
	cc.Cipher = cipher
	cc.Conn = conn
	return
}

func (cc *cipherConn) Read(buf []byte) (n int, err error) {

	n, err = cc.Conn.Read(buf)
	if n != 0 {
		cc.Decrypt(buf[:n], buf[:n])
	}
	return
}

func (cc *cipherConn) Write(buf []byte) (n int, err error) {
	tmp := make([]byte, len(buf))
	cc.Encrypt(tmp, buf)
	n, err = cc.Conn.Write(tmp)
	return
}
