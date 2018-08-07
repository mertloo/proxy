package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func debugGoroutine() {
	go func() {
		for _ = range time.Tick(3 * time.Second) {
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
	server := flag.String("server", "0.0.0.0:20001", "server addr (default: 0.0.0.0:20001)")
	nextserver := flag.String("nextserver", "", "next server addr")
	proto := flag.String("proto", "socks5", "server proto (default:socks5)")
	cryptmeth := flag.String("cryptmeth", "aes256cfb", "crypt meth (default:aes256cfb)")
	password := flag.String("password", "", "server password")
	pprof := flag.Bool("pprof", false, "start pprof server")

	flag.Parse()
	cfg := &Config{
		Addr:       *server,
		Proto:      *proto,
		NextServer: *nextserver,
		CryptMeth:  *cryptmeth,
		Password:   *password,
	}
	srv, err := NewServer(cfg)
	if err != nil {
		log.Println("new server err:", err)
		return
	}
	var addr string
	if *proto == "socks5" {
		addr = ":6060"
	} else {
		addr = ":6061"
	}
	if *pprof {
		log.Println("pp addr", addr)
		pprofRun(addr)
		debugGoroutine()
	}
	srv.ListenAndServe()
}
