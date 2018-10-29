package main

import (
	"flag"
)

func main() {
	debug := flag.Bool("debug", false, "enable pprof, goroNum for debug")
	upstream := flag.String("upstream", "socks5://127.0.0.1:1080", "upstream conn string")
	downstream := flag.String("downstream", "ssocks://aes256cfb(woshimima)@127.0.0.1:1990", "downstream conn string")

	flag.Parse()

	if *debug {
		pprofRun(*pprof)
		goroutineNum(*goro)
	}

	srv := new(Server)
	if err := parseConfig(upstream, downstream, srv); err != nil {
		panic(err)
	}
	srv.ListenAndServe()
}

func parseConfig(upstream, downstream string, server *Server) error {
	// e.g.
	// -u socks5://127.0.0.1:1080 -d ssocks://aes256cfb(woshimima)@127.0.0.1:1990

}
