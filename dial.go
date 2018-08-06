package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"strconv"
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
	cipher, err := info.newFunc(iv, sd.password)
	if err != nil {
		return
	}
	cc := newCipherConn(cipher, c)
	if err != nil {
		return
	}
	c = cc
	buf := make([]byte, 259)
	buf[0] = 0x03
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	alen := byte(len(host))
	buf[1] = alen
	copy(buf[2:2+alen], host)
	portn, err := strconv.Atoi(port)
	if err != nil {
		return
	}
	buf[2+alen] = byte(uint16(portn) >> 8)
	buf[3+alen] = byte(portn)
	c.Write(buf[:4+alen])
	c.Read(buf[:1])
	if buf[0] != 0x01 {
		err = fmt.Errorf("conn remote failed")
		return
	}
	conn = c
	return
}
