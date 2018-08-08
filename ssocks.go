package main

import (
	"crypto/rand"
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

// FIXME --
type decstreamConn struct {
	net.Conn
	*Cipher
}

func (dc *decstreamConn) Read(buf []byte) (int, error) {
	n, err := dc.Conn.Read(buf)
	if n != 0 {
		dc.Decrypt(buf[:n], buf[:n])
	}
	return n, err
}

type encstreamConn struct {
	net.Conn
	*Cipher
}

func (ec *encstreamConn) Write(buf []byte) (int, error) {
	tmp := make([]byte, len(buf))
	ec.Encrypt(tmp, buf)
	n, err := ec.Conn.Write(tmp)
	return n, err
}

func newDecstream(cryptMeth, password string, conn net.Conn) (net.Conn, error) {
	info, err := getCryptInfo(cryptMeth)
	if err != nil {
		return conn, err
	}
	buf := make([]byte, info.ivLen)
	n, err := conn.Read(buf)
	if n != info.ivLen || (err != nil && err != io.EOF) {
		err = fmt.Errorf("read iv err", buf[:n], err)
		return conn, err
	}
	cipher, err := info.newCipherFunc(buf, password)
	if err != nil {
		return conn, err
	}
	return &decstreamConn{conn, cipher}, nil
}

func newEncstream(cryptMeth, password string, conn net.Conn) (net.Conn, error) {
	info, err := getCryptInfo(cryptMeth)
	if err != nil {
		return conn, err
	}
	buf := make([]byte, info.ivLen)
	_, err = io.ReadFull(rand.Reader, buf)
	if err != nil {
		return conn, err
	}
	conn.Write(buf)
	cipher, err := info.newCipherFunc(buf, password)
	if err != nil {
		return conn, err
	}
	return &encstreamConn{conn, cipher}, nil
}

// -- FIXME

func (s *ssocks) setUpstream() (err error) {
	// read iv, init dec, read addr
	// write iv, init enc, conn remote, write ack
	s.upStream, err = newDecstream(s.server.cryptMeth, s.server.password, s.upStream)
	if err != nil {
		return
	}
	addr, err := s.readAddr()
	if err != nil {
		return
	}
	conn, err := s.server.dial(addr)
	if err != nil {
		return
	}
	s.downStream = newTimeoutConn(conn, s.timeout)
	s.upStream, err = newEncstream(s.server.cryptMeth, s.server.password, s.upStream)
	if err != nil {
		return
	}
	//s.upStream.Write([]byte{0x01})
	//--
	/*
		cipher, err := newCipher(s.server.cryptMeth, s.server.password, s.upStream)
		if err != nil {
			return
		}
		s.upStream = newCipherConn(cipher, s.upStream)
	*/
	return
}

func (s *ssocks) setDownstream() (err error) {
	/*
		addr, err := s.readAddr()
		fmt.Println("GOT ADDR", addr)
		if err != nil {
			return
		}
		conn, err := s.server.dial(addr)
		if err != nil {
			return
		}
		s.downStream = newTimeoutConn(conn, s.timeout)
		fmt.Println("CONN ADDR", addr)
		nw, err := s.upStream.Write([]byte{0x01})
		fmt.Println("ACKED", addr, nw, err)
	*/
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
