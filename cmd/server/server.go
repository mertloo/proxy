package main

import (
	"github.com/mertloo/proxy/ssocks"
)

func main() {
	srv := &ssocks.Server{"0.0.0.0:20002", "woshimima1234567"}
	srv.ListenAndServe()
}
