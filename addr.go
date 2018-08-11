package main

import (
	"fmt"
	"net"
	"strconv"
)

func ReadAddr(conn net.Conn) (addr string, err error) {
	buf := make([]byte, 259)
	n, e := conn.Read(buf[:2])
	if n != 2 || buf[0] != 0x03 || e != nil {
		err = fmt.Errorf("read addr head error (read: %v, err: %v)", buf[:n], e)
		return
	}
	hLen := int(buf[1])
	n, e = conn.Read(buf[:hLen+2])
	if n != hLen+2 || e != nil {
		err = fmt.Errorf("read addr error (read: %v, err: %v)", buf[:n], e)
		return
	}
	addr = fmt.Sprintf("%s:%d", buf[:hLen], uint16(buf[hLen])<<8|uint16(buf[hLen+1]))
	return
}

func WriteAddr(conn net.Conn, addr string) (err error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	portn, err := strconv.Atoi(port)
	if err != nil {
		return
	}
	buf := make([]byte, 259)
	buf[0] = 0x03
	hLen := len(host)
	buf[1] = byte(hLen)
	copy(buf[2:2+hLen], host)
	buf[2+hLen] = byte(uint16(portn) >> 8)
	buf[3+hLen] = byte(portn)
	_, err = conn.Write(buf[:4+hLen])
	return
}
