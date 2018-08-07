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
}

func checkConfig(cfg *Config) (err error) {
	ok := false
	for _, proto := range []string{"socks5", "ssocks"} {
		if cfg.Proto == proto {
			ok = true
			break
		}
	}
	if !ok {
		err = fmt.Errorf("not support proto")
	}
	return
}

type server struct {
	addr       string
	proto      string
	nextServer string
	cryptMeth  string
	password   string
	dialer
}

func NewServer(cfg *Config) (srv *server, err error) {
	err = checkConfig(cfg)
	if err != nil {
		return
	}
	srv = &server{
		addr:       cfg.Addr,
		nextServer: cfg.NextServer,
		cryptMeth:  cfg.CryptMeth,
		password:   cfg.Password,
		proto:      cfg.Proto,
	}
	switch srv.proto {
	case "socks5":
		srv.dialer = &ssocksDialer{srv.cryptMeth, srv.password, srv.nextServer}
	case "ssocks":
		srv.dialer = defaultDialer
	}
	return
}

func (srv *server) ListenAndServe() {
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		log.Println("listen err:", err)
		return
	}
	log.Println("listen at:", srv.addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept err:", err)
			continue
		}
		go srv.handle(conn)
	}
}

func (srv *server) handle(conn net.Conn) {
	proxy := srv.newProxy(conn)
	defer proxy.close()
	fs := []func() error{
		proxy.setUpstream,
		proxy.setDownstream,
		proxy.transport,
	}
	for _, f := range fs {
		if err := f(); err != nil {
			log.Println(err)
			return
		}
	}
}

func (srv *server) newProxy(conn net.Conn) (p proxy) {
	switch srv.proto {
	case "socks5":
		p = newSocks5(conn, srv)
	case "ssocks":
		p = newSSocks(conn, srv)
	}
	return
}
