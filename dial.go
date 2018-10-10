package main

import (
	"net"
	"time"
)

type TCPDialer struct {
	Timeout time.Duration
}

func (d *TCPDialer) Dial(addr string) (conn net.Conn, err error) {
	return net.DialTimeout("tcp", addr, d.Timeout)
}

type SSocksDialer struct {
	Server   string
	Password string
	Cipher   *cipherInfo
	Timeout  time.Duration
}

func (d *SSocksDialer) Dial(addr string) (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp", d.Server, d.Timeout)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			conn.Close()
			return
		}
	}()
	var ssocks SSocks
	eConn, err := ssocks.NewEConn(conn, d.Password, d.Cipher)
	if err != nil {
		return
	}
	conn = eConn
	err = WriteAddr(conn, addr)
	if err != nil {
		return
	}
	dConn, err := ssocks.NewDConn(conn, d.Password, d.Cipher)
	if err != nil {
		return
	}
	conn = dConn
	return
}
