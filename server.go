package main

import (
	"log"
	"net"
)

type ServeAgent interface {
	Serve()
}

type Server struct {
	Addr     string
	NewAgent func(net.Conn) ServeAgent
}

func (srv *Server) ListenAndServe() {
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
		agent := srv.NewAgent(conn)
		go agent.Serve()
	}
}
