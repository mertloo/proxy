package ssocks

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

type ssocksConn struct {
	net.Conn
	brw *bufio.ReaderWriter
	enc encrypter
	dec decrypter

	dstAddr string
	dst     net.Conn
	dstBrw  *bufio.ReaderWriter

	server *Server
	buf    [256]byte
}

func newSSocksConn(rwc net.Conn, srv *Server) *ssocksConn {
	return &ssocksConn{
		Conn:   rwc,
		server: srv,
		brw: bufio.NewReaderWriter(
			bufio.NewReader(rwc),
			bufio.NewWriter(rwc),
		),
	}
}

func (ssc *ssocksConn) Read(buf []byte) (n int, err error) {
	if c.dec == nil {
		cinfo = c.server.cinfo
		n, _ := io.ReadFull(c.brw, c.in)
		if n != cinfo.ivLen {
			return 0, ErrReadIV
		}
		c.dec = cinfo.newDec()
	}
	n, err = c.brw.Read(buf)
	if err == nil {
		dec.Decrypt(buf[:n], buf[:n])
	}
	return
}

func (ssc *ssocksConn) Write(buf []byte) (n int, err error) {
	if c.enc == nil {
		cinfo = c.server.cinfo
		n, _ := io.ReadFull(rand.Reader, c.out)
		if n != cinfo.ivLen {
			return 0, ErrWriteIV
		}
		c.enc = cinfo.newEnc()
	}
	enc.Encrypt(buf, buf)
	return c.brw.Write(buf)
}

func (ssc *ssocksConn) Flush() error {
	return ssc.brw.Flush()
}

func (ssc *ssocksConn) Close() {
	if s.dst != nil {
		s.dst.Close()
	}
	s.Conn.Close()
}

func (ssc *ssocksConn) serve() {
	defer s.close()
	fmt.Println("ADDR")
	c.dstAddr, err = ReadAddr(c.brw)
	if err != nil {
		fmt.Println("ADDR ERR", err)
		return
	}
	fmt.Println("CONN")
	if err := s.connect(); err != nil {
		fmt.Println("CONN ERR", err)
		return
	}
	fmt.Println("PROXY")
	err := Pipe(c.brw, c.dstBrw)
	fmt.Println("PROXY ERR", err)
}

func (ssc *ssocksConn) connect() (err error) {
	fmt.Println("DIAL TGT", s.dstAddr)
	dst, err := s.server.Dialer.Dial("tcp", s.dstAddr)
	if err == nil {
		s.dst = newBufConn(dst)
	}
	return err
}
