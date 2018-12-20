package ssocks

import (
	"log"
	"net"
	"time"

	"github.com/mertloo/proxy"
)

type Server struct {
	Addr     string
	Method   string
	Password string
	Timeout  time.Duration
	cinfo    *cipherInfo
}

func (srv *Server) ListenAndServe() {
	cinfo, err := GetCipherInfo(srv.Method, srv.Password)
	if err != nil {
		log.Println("get cipher err", err)
		return
	}
	srv.cinfo = cinfo
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Println("listen err", err)
		return
	}
	log.Println("listen at:", srv.Addr)
	for {
		rwc, err := ln.Accept()
		if err != nil {
			log.Println("accept err", err)
			continue
		}
		go srv.Handle(rwc)
	}
}

func (srv *Server) Handle(rwc net.Conn) {
	tc := &timeoutConn{Conn: rwc, Timeout: srv.Timeout}
	c := newConn(tc, srv.cinfo)
	srcAddr := c.RemoteAddr()
	log.Printf("%s ssocks addr.\n", srcAddr)
	dstAddr, err := proxy.ReadAddr(c)
	if err != nil {
		log.Printf("%s ssocks read addr err %s.\n", srcAddr, err)
		return
	}
	log.Printf("%s ssocks dial %s.\n", srcAddr, dstAddr)
	dst, err := net.DialTimeout("tcp", dstAddr, srv.Timeout)
	if err != nil {
		log.Printf("%s ssocks dial %s err %s.\n", srcAddr, dstAddr, err)
		return
	}
	dst = &timeoutConn{Conn: dst, Timeout: srv.Timeout}
	log.Printf("%v -> %v pipe closed.\n", srcAddr, dstAddr)
	err, rerr := proxy.Pipe(dst, c)
	if err != nil {
		log.Printf("%v -> %v ERROR %v.\n", srcAddr, dstAddr, err)
	}
	if rerr != nil {
		log.Printf("%v -> %v ERROR %v.\n", dstAddr, srcAddr, rerr)
	}
	return
}
