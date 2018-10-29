package main

import (
	"net"
	"time"
)

type Dialer interface {
	Dial(network, addr string) (conn net.Conn, err error)
}

type SSocksDialer struct {
	Server  string
	Timeout time.Duration

	Method   string
	Password string
	cinfo    *cipherInfo
	ready    bool
}

func (ssd *SSocksDialer) Dial(network, addr string) (net.Conn, error) {
	// TBD: set default
	if !ssd.ready {
		ssd.cinfo, err = getCipherInfo(srv.Method, srv.Password)
		if err != nil {
			return nil, err
		}
		ssd.ready = true
	}
	rwc, err := net.DialTimeout(network, ssd.Server, ssd.Timeout)
	if err != nil {
		return nil, err
	}
	// TBD: conn with srv/cln cipher
	ssc := newSSocksConn(rwc, nil)
	if err = WriteAddr(ssc, addr); err == nil {
		err = ssc.Flush()
	}
	if err != nil {
		ssc.Close()
		ssc = nil
	}
	return ssc, err
}
