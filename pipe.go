package proxy

import (
	"fmt"
	"io"
	"net"
)

type ErrTimeout interface {
	Error() string
	Timeout() bool
}

func Pipe(dst, src net.Conn) (err error) {
	ch := make(chan error)
	defer close(ch)
	var eu, ed error
	go func() {
		_, e := io.Copy(dst, src)
		ch <- e
	}()
	_, eu = io.Copy(src, dst)
	ed = <-ch
	for _, e := range []error{eu, ed} {
		if e == nil {
			continue
		}
		if _, ok := e.(ErrTimeout); !ok {
			err = fmt.Errorf("agent.transport() error. (eu: %v, ed: %v)", eu, ed)
			break
		}
	}
	return
}
