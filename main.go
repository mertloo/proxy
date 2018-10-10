package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func goroutineNum(n int) {
	go func() {
		for _ = range time.Tick(time.Duration(n) * time.Second) {
			log.Println("#goroutines", runtime.NumGoroutine())
		}
	}()
}

func pprofRun(addr string) {
	go func() {
		log.Println(http.ListenAndServe(addr, nil))
	}()
}

func main() {
	proto := flag.String("proto", "socks5", "server proto (socks5|ssocks default: socks5)")
	addr := flag.String("addr", "0.0.0.0:20001", "server addr (default: 0.0.0.0:20001)")
	next := flag.String("next", "", "next server addr")
	cipher := flag.String("cipher", "aes256cfb", "cipher name (default: aes256cfb)")
	password := flag.String("password", "", "ssocks password")
	timeoutN := flag.Int("timeout", 2, "dial/read/write timeout (default: 2s)")
	pprof := flag.String("pprof", "", "pprof http addr")
	goro := flag.Int("goro", 0, "goroutine num print interval (default: 0)")

	flag.Parse()

	if *pprof != "" {
		pprofRun(*pprof)
	}

	if *goro != 0 {
		goroutineNum(*goro)
	}

	timeout := time.Duration(*timeoutN) * time.Second
	var dialer Dialer
	switch *proto {
	case "socks5":
		dialer = &SSocksDialer{*next, *cipher, *password, timeout}
	case "ssocks":
		dialer = &TCPDialer{timeout}
	}

	srv := new(Server)
	srv.Addr = *addr
	srv.NewAgent = func(conn net.Conn) ServeAgent {
		return NewAgent(conn, *proto, *cipher, *password, dialer, timeout)
	}
	srv.ListenAndServe()

}
