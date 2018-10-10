package main

import (
	"bytes"
	"fmt"
	"net"
)

var (
	socks5NoAuthResp  = []byte{0x05, 0x00}
	socks5ConnectReq  = []byte{0x05, 0x01, 0x00}
	socks5ConnectResp = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10}
)

type Socks5 struct{}

func (s *Socks5) Auth(conn net.Conn) (err error) {
	buf := make([]byte, 257)
	n, e := conn.Read(buf)
	if n < 3 || buf[0] != 0x05 || buf[1] < 1 {
		err = fmt.Errorf("invalid socks5 auth request (req: %v, err: %v)", buf[:n], e)
		return
	}

	hasNoAuth := false
	for i, nmeth := 0, int(buf[1]); i < nmeth; i++ {
		if buf[2+i] == 0x00 {
			hasNoAuth = true
			break
		}
	}
	if !hasNoAuth {
		err = fmt.Errorf("no NoAuth method in socks5 auth request")
		return
	}

	n, e = conn.Write(socks5NoAuthResp)
	if n != len(socks5NoAuthResp) || e != nil {
		err = fmt.Errorf("write socks5 noauth error (write: %d, err: %v)", n, e)
	}
	return
}

func (s *Socks5) Connect(conn net.Conn, dialer Dialer) (dstConn net.Conn, err error) {
	buf := make([]byte, 3)
	n, e := conn.Read(buf[:3])
	if n != 3 || !bytes.Equal(buf[:3], socks5ConnectReq) {
		err = fmt.Errorf("read socks5 connect cmd error (read: %v, err: %v)", buf[:n], e)
		return
	}
	defer func() {
		if err != nil {
			return
		}
		n, e = conn.Write(socks5ConnectResp)
		if n != len(socks5ConnectResp) || e != nil {
			err = fmt.Errorf("write socks5 connect cmd error (write: %v, err: %v)", buf[:n], e)
			return
		}
	}()
	addr, err := ReadAddr(conn)
	if err != nil {
		return
	}
	dstConn, err = dialer.Dial(addr)
	return
}
