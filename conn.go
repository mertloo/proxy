package main

import (
	"net"
	"time"
)

type timeoutConn struct {
	net.Conn
	timeout time.Duration
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

type EConn struct {
	net.Conn
	Encrypter
}

func (ec *EConn) Write(buf []byte) (n int, err error) {
	wb := make([]byte, len(buf))
	ec.Encrypt(wb, buf)
	n, err = ec.Conn.Write(wb)
	return
}

type DConn struct {
	net.Conn
	Decrypter
}

func (dc *DConn) Read(buf []byte) (n int, err error) {
	n, err = dc.Conn.Read(buf)
	if err == nil {
		dc.Decrypt(buf[:n], buf[:n])
	}
	return
}
