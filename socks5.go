package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	noAuth    = []byte{0x05, 0x00}
	connected = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10}
	authReq   = []byte{0x05, 0x01, 0x00}
	connReq   = []byte{0x05, 0x01, 0x00}

	authReqErr    = fmt.Errorf("auth req err")
	connReqErr    = fmt.Errorf("conn req err")
	notSupportErr = fmt.Errorf("not support err")
)

type socks5 struct {
	upStream   net.Conn
	downStream net.Conn
	timeout    time.Duration
	server     *server
}

func newSocks5(conn net.Conn, srv *server) proxy {
	s := new(socks5)
	s.timeout = 2 * time.Second
	s.upStream = newTimeoutConn(conn, s.timeout)
	s.server = srv
	return s
}

func (s *socks5) setUpstream() error {
	return s.handshake()
}

func (s *socks5) setDownstream() (err error) {
	conn, err := s.connect()
	if err == nil {
		s.downStream = newTimeoutConn(conn, s.timeout)
	}
	return
}

func (s *socks5) transport() error {
	return trans(s.upStream, s.downStream)
}

func (s *socks5) close() error {
	return pclose(s.upStream, s.downStream)
}

func (s *socks5) handshake() error {
	buf := make([]byte, 3)
	n, _ := s.upStream.Read(buf)
	if !bytes.Equal(buf[:n], authReq) {
		log.Println("handshake", buf[:n])
		return authReqErr
	}
	_, err := s.upStream.Write(noAuth)
	return err
}

func (s *socks5) connect() (net.Conn, error) {
	buf := make([]byte, 259)
	n, _ := s.upStream.Read(buf)
	if n == 0 {
		return nil, connReqErr
	}
	m := len(connReq)
	if !bytes.Equal(buf[:m], connReq) {
		return nil, connReqErr
	}
	atyp := buf[m]
	var addr string
	switch atyp {
	case 0x03:
		alen := buf[m+1]
		buf = buf[m+2:]
		host, port := buf[:alen], uint16(buf[alen])<<8|uint16(buf[alen+1])
		addr = fmt.Sprintf("%s:%d", host, port)
	default:
		return nil, notSupportErr
	}
	conn, err := s.server.dial(addr)
	if err != nil {
		return nil, err
	}
	_, err = s.upStream.Write(connected)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
