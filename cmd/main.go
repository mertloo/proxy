package main

import (
	"flag"
	"log"

	"github.com/mertloo/proxy"
	"github.com/mertloo/proxy/socks5"
	"github.com/mertloo/proxy/ssocks"
)

func main() {
	server := flag.String("server", "socks5://127.0.0.1:1080", "server type")
	dialer := flag.String("dialer", "ssocks://aes256cfb:woshimima@0.0.0.0:8388", "dialer type")
	debug := flag.Bool("debug", false, "open debug")
	pprof := flag.String("pprof", "0.0.0.0:6088", "pprof http addr on debug")
	goro := flag.Int("goro", 5, "goroNum print second interval on debug")
	flag.Parse()

	if *debug {
		proxy.PProfRun(*pprof)
		proxy.GoroNum(*goro)
	}

	sconf, err := proxy.ParseConfig(*server)
	if err != nil {
		log.Println(err)
		return
	}
	switch sconf.Proto {
	case "socks5":
		dconf, err := proxy.ParseConfig(*dialer)
		if err != nil {
			log.Println(err)
			return
		}
		srv := &socks5.Server{Addr: sconf.Addr}
		if dconf.Proto == "ssocks" {
			srv.Dialer = &ssocks.Dialer{
				Server:   dconf.Addr,
				Method:   dconf.Method,
				Password: dconf.Password,
			}
		}
		srv.ListenAndServe()
	case "ssocks":
		srv := &ssocks.Server{
			Addr:     sconf.Addr,
			Method:   sconf.Method,
			Password: sconf.Password,
		}
		srv.ListenAndServe()
	default:
		log.Println("invalid server type %s", sconf.Proto)
	}
	return
}
