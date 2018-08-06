package main

import (
	"fmt"
	"log"
	"net"
)

type Config struct {
	Addr       string
	Proto      string
	NextServer string
	CryptMeth  string
	Password   string
	Debug      bool
}

type conn struct {
	proxy
}

func (c *conn) serve() {
	defer c.close()
	fs := []func() error{c.setUpstream, c.setDownstream, c.transport}
	for _, f := range fs {
		if err := f(); err != nil {
			log.Println("serve err:", err)
			return
		}
	}
}

type server struct {
	addr       string
	proto      string
	nextServer string
	cryptMeth  string
	password   string
	debug      bool
	newConn    func(net.Conn, *server) proxy
	dialer
}

func NewServer(cfg *Config) (*server, error) {
	srv := new(server)
	srv.addr = cfg.Addr
	srv.nextServer = cfg.NextServer
	srv.cryptMeth = cfg.CryptMeth
	srv.password = cfg.Password
	err := fmt.Errorf("not support proto")
	for _, proto := range []string{"socks5", "ssocks"} {
		if cfg.Proto == proto {
			srv.proto = cfg.Proto
			err = nil
			break
		}
	}
	if err != nil {
		return nil, err
	}
	srv.debug = cfg.Debug
	switch srv.proto {
	case "socks5":
		srv.newConn = newSocks5
		srv.dialer = &ssocksDialer{srv.cryptMeth, srv.password, srv.nextServer}
	case "ssocks":
		srv.newConn = newSSocks
		srv.dialer = defaultDialer
	}
	return srv, nil
}

func (srv *server) ListenAndServe() {
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		log.Println("listen err:", err)
		return
	}
	log.Println("listen at:", srv.addr)
	for {
		tcpConn, err := ln.Accept()
		if err != nil {
			log.Println("accept err:", err)
			continue
		}
		proxy := srv.newConn(tcpConn, srv)
		c := &conn{proxy}
		go c.serve()
	}
}
