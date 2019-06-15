package ssocks

import (
	"net"
	"time"

	"github.com/mertloo/proxy"
)

type dialer struct {
	server   string
	method   string
	password string
	timeout  time.Duration
	cinfo    *cipherInfo
}

func NewDialer(server, method, password string,
	timeout time.Duration) (d *dialer, err error) {
	cinfo, err := getCipherInfo(method, password)
	if err != nil {
		return
	}
	d = &dialer{
		server:   server,
		method:   method,
		password: password,
		cinfo:    cinfo,
	}
	return
}

func (d *dialer) Dial(network, addr string) (c net.Conn, err error) {
	rwc, err := net.DialTimeout(network, d.server, d.timeout)
	if err != nil {
		return
	}
	tc := &proxy.TimeoutConn{Conn: rwc, Timeout: d.timeout}
	c = newConn(tc, d.cinfo)
	if err = proxy.WriteAddr(c, addr); err != nil {
		c.Close()
		c = nil
	}
	return
}
