package socks5

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

var (
	NoAuthResp     = []byte{0x05, 0x00}
	CmdConnect     = []byte{0x05, 0x01, 0x00}
	CmdConnectResp = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10}
	ErrVerion      = errors.New("socks5 version error")
	ErrMethod      = errors.New("socks5 method error")
	ErrConnect     = errors.New("socks5 connect error")
)

type socks5Conn struct {
	net.Conn
	brw *bufio.ReaderWriter

	dstAddr string
	dst     net.Conn
	dstBrw  *bufio.ReaderWriter

	server *Server
	buf    [256]byte
}

func newSocks5Conn(rwc net.Conn, srv *Server) *socks5Conn {
	return &socks5Conn{
		Conn:   rwc,
		server: srv,
		brw: bufio.NewReaderWriter(
			bufio.NewReader(rwc),
			bufio.NewWriter(rwc),
		),
	}
}

func (sc *socks5Conn) serve() {
	defer s.close()
	fmt.Println("AUTH")
	if err := s.auth(); err != nil {
		fmt.Println("AUTH ERR", err)
		return
	}
	fmt.Println("CONN")
	if err := s.connect(); err != nil {
		fmt.Println("CONN ERR", err)
		return
	}
	fmt.Println("PROXY")
	err := Pipe(s.dst, s.src)
	fmt.Println("PROXY ERR", err)
}

func (sc *socks5Conn) close() {
	if s.dst != nil {
		s.dst.Close()
	}
	s.src.Close()
}

func (sc *socks5Conn) auth() (err error) {
	_, err = io.ReadFull(s.src, s.buf[:2])
	if err != nil {
		return err
	}
	ver, nmeth := s.buf[0], int(s.buf[1])
	if ver != 0x05 {
		return ErrVerion
	}
	_, err = io.ReadFull(s.src, s.buf[:nmeth])
	if err != nil {
		return err
	}
	for i := 0; i < nmeth; i++ {
		if s.buf[i] == 0x00 {
			_, err = s.src.Write(NoAuthResp)
			if err == nil {
				err = s.src.Flush()
			}
			return
		}
	}
	return ErrMethod
}

func (sc *socks5Conn) connect() (err error) {
	fmt.Println("READ CONN CMD")
	_, err = io.ReadFull(s.src, s.buf[:3])
	if err != nil {
		return err
	}
	fmt.Println("CHECK CONN CMD")
	if !bytes.Equal(s.buf[:3], CmdConnect) {
		return ErrConnect
	}
	fmt.Println("READ ADDR")
	s.dstAddr, err = ReadAddr(s.src)
	if err != nil {
		return
	}
	fmt.Println("DIAL TGT", s.dstAddr)
	dst, err := s.dialer.Dial("tcp", s.dstAddr)
	if err != nil {
		return
	}
	s.dst = newBufConn(dst)
	_, err = s.src.Write(CmdConnectResp)
	fmt.Println("RESP CONN CMD", CmdConnectResp)
	if err == nil {
		err = s.src.Flush()
	}
	return err
}
