package main

import (
	"github.com/mertloo/proxy/socks5"
	"github.com/mertloo/proxy/ssocks"
)

func main() {
	d := &ssocks.Dialer{"0.0.0.0:20002", "woshimima1234567"}
	srv := &socks5.Server{"0.0.0.0:20001", d}
	srv.ListenAndServe()
}
