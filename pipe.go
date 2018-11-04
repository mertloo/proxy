package proxy

import (
	"io"
	"net"
)

type ErrTimeout interface {
	Error() string
	Timeout() bool
}

func Pipe(dst, src net.Conn) (err, rerr error) {
	ch := make(chan struct{})
	go func() {
		_, err = io.Copy(dst, src)
		close(ch)
	}()
	_, rerr = io.Copy(src, dst)
	<-ch
	if err != nil {
		if _, ok := err.(ErrTimeout); ok {
			err = nil
		}
	}
	if rerr != nil {
		if _, ok := rerr.(ErrTimeout); ok {
			rerr = nil
		}
	}
	return
}
