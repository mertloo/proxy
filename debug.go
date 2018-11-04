package proxy

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func GoroNum(n int) {
	go func() {
		for _ = range time.Tick(time.Duration(n) * time.Second) {
			log.Println("#goroutines", runtime.NumGoroutine())
		}
	}()
}

func PProfRun(addr string) {
	go func() {
		log.Println(http.ListenAndServe(addr, nil))
	}()
}
