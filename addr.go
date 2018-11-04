package proxy

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
	var buf [256]byte
	_, err = conn.Read(buf[:1])
	if err != nil {
		return
	}
	switch buf[0] {
	case IPv4:
		_, err = conn.Read(buf[:6])
		if err != nil {
			return
		}
		port := uint16(buf[4])<<8 | uint16(buf[5])
		addr = fmt.Sprintf("%s:%d", net.IPv4(buf[0], buf[1], buf[2], buf[3]), port)
	case FQDN:
		_, err = conn.Read(buf[:1])
		if err != nil {
			return
		}
		alen := int(buf[0])
		_, err = conn.Read(buf[:alen+2])
		if err != nil {
			return
		}
		port := uint16(buf[alen])<<8 | uint16(buf[alen+1])
		addr = fmt.Sprintf("%s:%d", buf[:alen], port)
	default:
		err = fmt.Errorf("not support atyp %v", buf[0])
	}
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
