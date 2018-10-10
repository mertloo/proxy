package main

import (
	"fmt"
	"net"
	"strconv"
)

const (
	IPv4 = 0x01
	FQDN = 0x03
)

func ReadAddr(conn net.Conn) (addr string, err error) {
	buf := make([]byte, 259)
	n, e := conn.Read(buf)
	if buf[0] == IPv4 && n == 7 {
		port := uint16(buf[n-2])<<8 | uint16(buf[n-1])
		addr = fmt.Sprintf("%s:%d", net.IPv4(buf[1], buf[2], buf[3], buf[4]), port)
		return
	}
	if buf[0] == FQDN && n == int(buf[1])+4 {
		port := uint16(buf[n-2])<<8 | uint16(buf[n-1])
		host := string(buf[2 : 2+int(buf[1])])
		addr = fmt.Sprintf("%s:%d", host, port)
		return
	}
	err = fmt.Errorf("read address error (buf: %v, err: %v)", buf[:n], e)
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
