package main

import (
	"fmt"
	"io"
	"net"
)

type proxy interface {
	setUpstream() error
	setDownstream() error
	transport() error
	close() error
}

func trans(upStream, downStream net.Conn) error {
	c := make(chan error)
	go func(c chan error) {
		_, err := io.Copy(upStream, downStream)
		c <- err
	}(c)
	_, eo := io.Copy(downStream, upStream)
	ei := <-c
	if eo != nil || ei != nil {
		return fmt.Errorf("ei: %s, eo: %s", ei, eo)
	}
	return nil
}
func pclose(upStream, downStream net.Conn) error {
	var eu, ed error
	eu = upStream.Close()
	if downStream != nil {
		ed = downStream.Close()
	}
	if eu != nil || ed != nil {
		return fmt.Errorf("eu: %v, ed: %v\n", eu, ed)
	}
	return nil
}
