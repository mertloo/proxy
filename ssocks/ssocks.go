package ssocks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	ssocksAtypHost = 0x03
)

type Dialer struct {
	ServerAddr string
	Password   string
}

func (d *Dialer) Dial(addr string) (io.ReadWriteCloser, error) {
	rwc, err := net.Dial("tcp", d.ServerAddr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil && rwc != nil {
			rwc.Close()
		}
	}()
	block, err := aes.NewCipher([]byte(d.Password))
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	n, err := io.ReadFull(rand.Reader, iv)
	if n != aes.BlockSize || err != nil {
		return nil, err
	}
	c := &conn{
		rwc:           rwc,
		remoteAddr:    addr,
		encryptStream: cipher.NewCFBEncrypter(block, iv),
		decryptStream: cipher.NewCFBDecrypter(block, iv),
	}
	n, err = c.rwc.Write(iv)
	if err != nil {
		return nil, err
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 269)
	buf[0], buf[1] = ssocksAtypHost, uint8(len(host))
	copy(buf[2:2+len(host)], host)
	portn, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	binary.BigEndian.PutUint16(buf[2+len(host):4+len(host)], uint16(portn))
	n, err = c.Write(buf[:4+len(host)])
	if err != nil {
		return nil, err
	}
	return c, nil
}

type conn struct {
	server        *Server
	rwc           net.Conn
	remoteAddr    string
	remoteConn    net.Conn
	encryptStream cipher.Stream
	decryptStream cipher.Stream
}

func (c *conn) Read(buf []byte) (n int, err error) {
	n, err = c.rwc.Read(buf)
	if n == 0 {
		return
	}
	c.decryptStream.XORKeyStream(buf[:n], buf[:n])
	return
}

func (c *conn) Write(buf []byte) (n int, err error) {
	ciphertext := make([]byte, len(buf))
	c.encryptStream.XORKeyStream(ciphertext, buf)
	n, err = c.rwc.Write(ciphertext)
	return
}

func (c *conn) serve() {
	defer c.Close()
	block, err := aes.NewCipher([]byte(c.server.Password))
	if err != nil {
		return
	}
	iv := make([]byte, aes.BlockSize)
	n, err := c.rwc.Read(iv)
	if n != aes.BlockSize || err != nil {
		return
	}
	c.encryptStream = cipher.NewCFBEncrypter(block, iv)
	c.decryptStream = cipher.NewCFBDecrypter(block, iv)
	buf := make([]byte, 256)
	n, err = c.Read(buf[:1])
	if n != 1 || err != nil {
		return
	}
	switch buf[0] {
	case ssocksAtypHost:
		n, err = c.Read(buf[:1])
		if n != 1 || err != nil {
			return
		}
	default:
		return
	}
	alen := buf[0]
	n, err = c.Read(buf[:alen+2])
	if n != int(alen)+2 || err != nil {
		return
	}
	host, port := buf[:alen], uint16(buf[alen])<<8|uint16(buf[alen+1])
	c.remoteAddr = fmt.Sprintf("%s:%d", host, port)
	c.remoteConn, err = net.Dial("tcp", c.remoteAddr)
	if err != nil {
		return
	}
	defer c.remoteConn.Close()
	quit := make(chan struct{})
	go c.pipe(c, c.remoteConn, quit)
	c.pipe(c.remoteConn, c, quit)
	return

}

func (c *conn) pipe(dst io.Writer, src io.Reader, quit chan struct{}) {
	for {
		_, err := io.Copy(dst, src)
		select {
		case <-quit:
			return
		default:
			if err != nil {
				close(quit)
				return
			}
		}
	}
}

func (c *conn) Close() error {
	return c.rwc.Close()
}

type Server struct {
	Addr     string
	Password string
}

func (srv *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(ln net.Listener) error {
	defer ln.Close()
	for {
		rwc, err := ln.Accept()
		if err != nil {
			continue
		}
		c := srv.newConn(rwc)
		go c.serve()
	}
}

func (srv *Server) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
	}
	return c
}
