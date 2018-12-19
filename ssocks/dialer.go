package ssocks

import (
	"net"

	"github.com/mertloo/proxy"
)

type Dialer struct {
	Server   string
	Method   string
	Password string
	cinfo    *cipherInfo
}

func (d *Dialer) Dial(network, addr string) (c net.Conn, err error) {
	if d.cinfo == nil {
		d.cinfo, err = GetCipherInfo(d.Method, d.Password)
		if err != nil {
			return
		}
	}
	rwc, err := net.Dial(network, d.Server)
	if err != nil {
		return
	}
	c = newConn(rwc, d.cinfo)
	if err = proxy.WriteAddr(c, addr); err != nil {
		c.Close()
		c = nil
	}
	return
}
