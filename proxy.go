package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

type proxy interface {
	setUpstream() error
	setDownstream() error
	transport() error
	close() error
}

type timeoutErr interface {
	Error() string
	Timeout() bool
}

func trans(upStream, downStream net.Conn) error {
	c := make(chan error)
	go func(c chan error) {
		_, err := io.Copy(upStream, downStream)
		c <- err
	}(c)
	_, eo := io.Copy(downStream, upStream)
	ei := <-c
	var hasErr, isTimeoutErr bool
	for _, e := range []error{ei, eo} {
		if e != nil {
			if _, isTimeoutErr = e.(timeoutErr); isTimeoutErr {
				continue
			}
			hasErr = true
		}
	}
	if hasErr {
		return fmt.Errorf("ei: %v, eo: %v", ei, eo)
	}
	return nil
}

func pclose(upStream, downStream net.Conn) error {
	var eu, ed error
	eu = upStream.Close()
	if downStream != nil {
		ed = downStream.Close()
	}
	if eu != nil || ed != nil {
		return fmt.Errorf("eu: %v, ed: %v\n", eu, ed)
	}
	return nil
}

func parseAddr(buf []byte) (addr string, err error) {
	n := len(buf)
	if n < 5 {
		err = fmt.Errorf("invalid addr buff")
		return
	}
	atyp, hlen := buf[0], buf[1]
	if int(4+hlen) != n {
		err = fmt.Errorf("invalid addr buff")
		return
	}
	switch atyp {
	case 0x03:
		host := buf[2 : 2+hlen]
		port := uint16(buf[n-2])<<8 | uint16(buf[n-1])
		addr = fmt.Sprintf("%s:%d", host, port)
	default:
		err = fmt.Errorf("not support atyp")
		return
	}
	return
}

func formatAddr(buf []byte, addr string) (n int, err error) {
	if len(buf) < 5 {
		err = fmt.Errorf("short buff err")
		return
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	portn, err := strconv.Atoi(port)
	if err != nil {
		return
	}
	buf[0] = 0x03
	hlen := len(host)
	buf[1] = byte(hlen)
	copy(buf[2:2+hlen], host)
	buf[2+hlen] = byte(uint16(portn) >> 8)
	buf[3+hlen] = byte(portn)
	n = 4 + hlen
	return
}
