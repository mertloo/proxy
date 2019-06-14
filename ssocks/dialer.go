package ssocks

import (
	"net"
	"time"

	"github.com/mertloo/proxy"
)

type Dialer struct {
	Server   string
	Method   string
	Password string
	Timeout  time.Duration
	cinfo    *cipherInfo
}

func (d *Dialer) Dial(network, addr string) (c net.Conn, err error) {
	if d.cinfo == nil {
		d.cinfo, err = GetCipherInfo(d.Method, d.Password)
		if err != nil {
			return
		}
	}
	rwc, err := net.DialTimeout(network, d.Server, d.Timeout)
	if err != nil {
		return
	}
	tc := &proxy.TimeoutConn{Conn: rwc, Timeout: d.Timeout}
	c = newConn(tc, d.cinfo)
	if err = proxy.WriteAddr(c, addr); err != nil {
		c.Close()
		c = nil
	}
	return
}
