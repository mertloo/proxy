package proxy

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	zero     int64         = 0
	duration time.Duration = time.Second
)

var (
	units = []string{"B", "KB", "MB", "GB"}
)

type Stats struct {
	m              *sync.Map
	LastRecvBytes  int64
	LastTransBytes int64
}

func NewStats() *Stats {
	return &Stats{m: &sync.Map{}}
}

func (s *Stats) AddStats(sc net.Conn) {
	s.m.Store(sc, nil)
}

func (s *Stats) DelStats(sc net.Conn) {
	s.m.Delete(sc)
}

func (s *Stats) DoStats() {
	for _ = range time.Tick(duration) {
		var recvBytes, transBytes int64 = 0, 0
		s.m.Range(func(k, _ interface{}) bool {
			sc := k.(*StatsConn)
			recvBytes += atomic.SwapInt64(&sc.RecvBytes, zero)
			transBytes += atomic.SwapInt64(&sc.TransBytes, zero)
			return true
		})
		s.printLog(recvBytes, transBytes)
		s.LastRecvBytes, s.LastTransBytes = recvBytes, transBytes
	}
}

func (s *Stats) printLog(recvBytes, transBytes int64) {
	if s.LastRecvBytes == 0 && s.LastTransBytes == 0 &&
		s.LastRecvBytes == recvBytes && s.LastTransBytes == transBytes {
		return
	}
	convUnit := func(v int64) (int64, string) {
		i := 0
		for {
			tmpv, tmpi := v>>10, i+1
			if tmpv == 0 || tmpi == len(units) {
				break
			}
			v, i = tmpv, tmpi
		}
		return v, units[i]
	}
	vr, ur := convUnit(recvBytes)
	vt, ut := convUnit(transBytes)
	log.Printf("Received: %d %s/s	Transmitted: %d %s/s\n", vr, ur, vt, ut)
}

type StatsConn struct {
	net.Conn
	RecvBytes  int64
	TransBytes int64
}

func (sc *StatsConn) Read(buf []byte) (n int, err error) {
	n, err = sc.Conn.Read(buf)
	atomic.AddInt64(&sc.RecvBytes, int64(n))
	return n, err
}

func (sc *StatsConn) Write(buf []byte) (n int, err error) {
	n, err = sc.Conn.Write(buf)
	atomic.AddInt64(&sc.TransBytes, int64(n))
	return n, err
}
