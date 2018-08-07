package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type ssocks struct {
	upStream   net.Conn
	downStream net.Conn
	timeout    time.Duration
	server     *server
}

func newSSocks(conn net.Conn, srv *server) (s *ssocks) {
	s = &ssocks{
		timeout:  defaultTimeout,
		upStream: newTimeoutConn(conn, defaultTimeout),
		server:   srv,
	}
	return
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

func (s *ssocks) transport() error {
	return trans(s.upStream, s.downStream)
}

func (s *ssocks) close() error {
	return pclose(s.upStream, s.downStream)
}

func (s *ssocks) readAddr() (addr string, err error) {
	buf := make([]byte, 259)
	n, err := s.upStream.Read(buf)
	if n == 0 || (err != nil && err != io.EOF) {
		err = fmt.Errorf("read addr err buf, err: %v, %v", buf[:n], err)
		return
	}
	return parseAddr(buf[:n])
}
