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
		// TBD: err instead panic
		d.cinfo = GetCipherInfo(d.Method, d.Password)
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
