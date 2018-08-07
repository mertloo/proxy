package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"time"
)

var (
	defaultTimeout = 2 * time.Second
	defaultDialer  = new(tcpDialer)
)

type dialer interface {
	dial(addr string) (net.Conn, error)
}

type tcpDialer struct{}

func (td *tcpDialer) dial(addr string) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, defaultTimeout)
}

type ssocksDialer struct {
	cryptMeth  string
	password   string
	serverAddr string
}

func (sd *ssocksDialer) dial(addr string) (conn net.Conn, err error) {
	c, err := net.DialTimeout("tcp", sd.serverAddr, defaultTimeout)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			c.Close()
		}
	}()
	info, err := getCryptInfo(sd.cryptMeth)
	if err != nil {
		return
	}
	iv := make([]byte, info.ivLen)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return
	}
	c.Write(iv)
	cipher, err := info.newCipherFunc(iv, sd.password)
	if err != nil {
		return
	}
	c = newCipherConn(cipher, c)
	buf := make([]byte, 259)
	n, err := formatAddr(buf, addr)
	if err != nil {
		return
	}
	c.Write(buf[:n])
	n, err = c.Read(buf[:1])
	if n != 1 || buf[0] != 0x01 {
		err = fmt.Errorf("dial remote failed (n, buf, err: %v, %v, %v)", n, buf[:n], err)
		return
	}
	conn = c
	return
}
