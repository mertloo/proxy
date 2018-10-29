package main

import (
	"log"
	"net"
	"time"
)

var (
	defaultTimeout  = 10 * time.Second
	defaultDialer   = &net.Dialer{Timeout: defaultTimeout}
	defaultMethod   = "aes256cfb"
	defaultPassword = "woshimima"
)

type Server struct {
	Addr    string
	Dialer  Dialer
	Timeout time.Duration

	Shadow   bool
	Method   string
	Password string

	cinfo   *cipherInfo
	newConn func(net.Conn, *Server) interface {
		serve()
	}
}

func (srv *Server) ListenAndServe() {
	if err := srv.prepare(); err != nil {
		log.Println("prepare err:", err)
		return
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Println("listen err:", err)
		return
	}
	log.Println("listen at:", srv.Addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept err:", err)
			continue
		}
		c := srv.newConn(conn)
		go c.serve()
	}
}

func (srv *Server) prepare() error {
	if srv.Addr == "" {
		srv.Addr = "0.0.0.0:1990"
	}
	if srv.dialer == nil {
		srv.Dialer = defaultDialer
	}
	if srv.Timeout == 0 {
		srv.Timeout = defaultTimeout
	}
	if !srv.Shadow {
		srv.newConn = newSocks5Conn
		return nil
	}
	srv.newConn = newSSocksConn
	if srv.Method == "" {
		srv.Method = defaultMethod
	}
	if srv.Password == "" {
		srv.Password = defaultPassword
	}
	srv.cinfo, err = getCipherInfo(srv.Method, srv.Password)
	return err
}
