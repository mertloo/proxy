package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type ErrTimeout interface {
	Error() string
	Timeout() bool
}

type Dialer interface {
	Dial(addr string) (conn net.Conn, err error)
}

type Agent struct {
	upstream   net.Conn
	downstream net.Conn
	proto      string
	info       *cipherInfo
	timeout    time.Duration
	Dialer
}

func NewAgent(conn net.Conn, proto string, info *cipherInfo, dialer Dialer, timeout time.Duration) *Agent {
	agent := new(Agent)
	agent.timeout = timeout
	agent.upstream = &timeoutConn{conn, agent.timeout}
	agent.proto = proto
	agent.info = info
	agent.Dialer = dialer
	return agent
}

func (agent *Agent) Serve() {
	defer agent.close()
	fns := []func() error{
		agent.handshake,
		agent.transport,
	}
	for _, fn := range fns {
		if err := fn(); err != nil {
			log.Println(err)
			return
		}
	}
}

func (agent *Agent) close() {
	var eu, ed error
	eu = agent.upstream.Close()
	if agent.downstream != nil {
		ed = agent.downstream.Close()
	}
	if eu != nil || ed != nil {
		log.Printf("agent.close() error. (eu: %v, ed: %v)\n", eu, ed)
	}
	return
}

func (agent *Agent) handshake() (err error) {
	switch agent.proto {
	case "socks5":
		err = agent.socks5Handshake()
	case "ssocks":
		err = agent.ssocksHandshake()
	}
	return
}

func (agent *Agent) transport() (err error) {
	ch := make(chan error)
	defer close(ch)
	var eu, ed error
	go func() {
		_, e := io.Copy(agent.downstream, agent.upstream)
		ch <- e
	}()
	_, eu = io.Copy(agent.upstream, agent.downstream)
	ed = <-ch
	var hasErr bool
	for _, e := range []error{eu, ed} {
		if e == nil {
			continue
		}
		if _, ok := e.(ErrTimeout); !ok {
			hasErr = true
			break
		}
	}
	if hasErr {
		err = fmt.Errorf("agent.transport() error. (eu: %v, ed: %v)", eu, ed)
	}
	return
}

func (agent *Agent) socks5Handshake() (err error) {
	var socks5 Socks5
	err = socks5.Auth(agent.upstream)
	if err != nil {
		return
	}
	conn, err := socks5.Connect(agent.upstream, agent.Dialer)
	if err == nil {
		agent.downstream = &timeoutConn{conn, agent.timeout}
	}
	return
}

func (agent *Agent) ssocksHandshake() (err error) {
	var ssocks SSocks
	dConn, err := ssocks.NewDConn(agent.upstream, agent.info)
	if err != nil {
		return
	}
	agent.upstream = dConn

	conn, err := ssocks.Connect(agent.upstream, agent.Dialer)
	if err != nil {
		return
	}
	agent.downstream = &timeoutConn{conn, agent.timeout}

	eConn, err := ssocks.NewEConn(agent.upstream, agent.info)
	if err != nil {
		return
	}
	agent.upstream = eConn
	return
}
