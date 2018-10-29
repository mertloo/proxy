package main

import (
	"log"
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
