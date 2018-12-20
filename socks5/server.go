package socks5

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/mertloo/proxy"
)

var (
	NoAuthResp     = []byte{0x05, 0x00}
	CmdConnect     = []byte{0x05, 0x01, 0x00}
	CmdConnectResp = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10}
)

type Dialer interface {
	Dial(network, addr string) (net.Conn, error)
}

type Server struct {
	Addr    string
	Timeout time.Duration
	Dialer
}

func (srv *Server) ListenAndServe() {
	if srv.Dialer == nil {
		srv.Dialer = &net.Dialer{Timeout: srv.Timeout}
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Println("listen err", err)
		return
	}
	log.Println("listen at:", srv.Addr)
	for {
		rwc, err := ln.Accept()
		if err != nil {
			log.Println("accept err", err)
			continue
		}
		c := srv.newConn(rwc)
		go c.serve()
	}
}

func (srv *Server) newConn(rwc net.Conn) *conn {
	tc := &timeoutConn{Conn: rwc, Timeout: srv.Timeout}
	return &conn{Conn: tc, Server: srv}
}

type conn struct {
	net.Conn
	*Server
	buf [256]byte

	dst     net.Conn
	dstAddr string
}

func (c *conn) serve() {
	defer c.Close()
	srcAddr := c.RemoteAddr()
	log.Printf("%v socks5 auth.\n", srcAddr)
	err := c.auth()
	if err != nil {
		log.Printf("%v socks5 auth err %v.\n", srcAddr, err)
		return
	}
	log.Printf("%v socks5 connect.\n", srcAddr)
	c.dst, err = c.connect()
	if err != nil {
		log.Printf("%v socks5 connect err %v.\n", srcAddr, err)
		return
	}
	defer c.dst.Close()
	log.Printf("%v -> %v pipe.\n", srcAddr, c.dstAddr)
	err, rerr := proxy.Pipe(c.dst, c)
	log.Printf("%v -> %v pipe closed.\n", srcAddr, c.dstAddr)
	if err != nil {
		log.Printf("%v -> %v ERROR %v.\n", srcAddr, c.dstAddr, err)
	}
	if rerr != nil {
		log.Printf("%v -> %v ERROR %v.\n", c.dstAddr, srcAddr, rerr)
	}
}

func (c *conn) auth() (err error) {
	_, err = c.Read(c.buf[:2])
	if err != nil {
		return
	}
	v, nm := c.buf[0], int(c.buf[1])
	if v != 0x05 {
		return fmt.Errorf("not support version %v", v)
	}
	_, err = c.Read(c.buf[:nm])
	for i := 0; i < nm; i++ {
		if c.buf[i] == 0x00 {
			_, err = c.Write(NoAuthResp)
			return
		}
	}
	return fmt.Errorf("not support method %v", c.buf[:nm])
}

func (c *conn) connect() (dst net.Conn, err error) {
	_, err = c.Read(c.buf[:3])
	if err != nil {
		return
	}
	if !bytes.Equal(c.buf[:3], CmdConnect) {
		return nil, fmt.Errorf("not connect cmd %v", c.buf[:3])
	}
	c.dstAddr, err = proxy.ReadAddr(c)
	if err != nil {
		return
	}
	dst, err = c.Dial("tcp", c.dstAddr)
	if err == nil {
		if _, ok := c.Dialer.(*net.Dialer); ok {
			dst = &timeoutConn{Conn: dst, Timeout: c.Timeout}
		}
		_, err = c.Write(CmdConnectResp)
		if err != nil {
			dst.Close()
			dst = nil
		}
	}
	return
}

type timeoutConn struct {
	net.Conn
	Timeout time.Duration
}

func (tc *timeoutConn) Read(buf []byte) (n int, err error) {
	if tc.Timeout > 0 {
		t := time.Now().Add(tc.Timeout)
		if err = tc.SetReadDeadline(t); err != nil {
			return
		}
	}
	n, err = tc.Conn.Read(buf)
	return
}

func (tc *timeoutConn) Write(buf []byte) (n int, err error) {
	if tc.Timeout > 0 {
		t := time.Now().Add(tc.Timeout)
		if err = tc.SetWriteDeadline(t); err != nil {
			return
		}
	}
	n, err = tc.Conn.Write(buf)
	return
}
