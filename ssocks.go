package main

import (
	"fmt"
	"net"
	"time"
)

type ssocks struct {
	upStream   net.Conn
	downStream net.Conn
	timeout    time.Duration
	server     *server
}

func newSSocks(conn net.Conn, srv *server) proxy {
	s := new(ssocks)
	s.timeout = 2 * time.Second
	s.upStream = newTimeoutConn(conn, s.timeout)
	s.server = srv
	return s
}

func (s *ssocks) setUpstream() (err error) {
	cipher, err := newCipher(s.server.cryptMeth, s.server.password, s.upStream)
	if err != nil {
		return
	}
	s.upStream = newCipherConn(cipher, s.upStream)
	return
}

func (s *ssocks) setDownstream() (err error) {
	addr, err := s.readAddr()
	if err != nil {
		return
	}
	conn, err := s.server.dial(addr)
	if err != nil {
		return
	}
	s.downStream = newTimeoutConn(conn, s.timeout)
	s.upStream.Write([]byte{0x01})
	return
}

func (s *ssocks) readAddr() (string, error) {
	buf := make([]byte, 259)
	n, err := s.upStream.Read(buf)
	if n < 6 {
		return "", fmt.Errorf("read addr err %v %v", buf[:n], err)
	}
	atyp, alen := buf[0], buf[1]
	var addr string
	switch atyp {
	case 0x03:
		if int(2+alen) != len(buf[2:n]) {
			return "", fmt.Errorf("addr with wrong alen")
		}
		host := buf[2 : 2+alen]
		port := uint16(buf[n-2])<<8 | uint16(buf[n-1])
		addr = fmt.Sprintf("%s:%d", host, port)
	default:
		return "", fmt.Errorf("not support atyp")
	}
	return addr, nil
}

func (s *ssocks) transport() error {
	return trans(s.upStream, s.downStream)
}

func (s *ssocks) close() error {
	return pclose(s.upStream, s.downStream)
}
