package ssocks

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

type conn struct {
	net.Conn
	cinfo    *cipherInfo
	enc, dec cipher.Stream
	buf      [256]byte
}

func newConn(rwc net.Conn, ci *cipherInfo) net.Conn {
	return &conn{Conn: rwc, cinfo: ci}
}

func (c *conn) Read(buf []byte) (n int, err error) {
	if c.dec == nil {
		if err = c.initDec(); err != nil {
			return
		}
	}
	n, err = c.Conn.Read(buf)
	if err == nil {
		c.dec.XORKeyStream(buf[:n], buf[:n])
	}
	return
}

func (c *conn) Write(buf []byte) (n int, err error) {
	if c.enc == nil {
		if err = c.initEnc(); err != nil {
			return
		}
	}
	c.enc.XORKeyStream(buf, buf)
	n, err = c.Conn.Write(buf)
	return
}

func (c *conn) initEnc() (err error) {
	iv := c.buf[:c.cinfo.ivLen]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return
	}
	c.enc = c.cinfo.newEncStream(c.cinfo.block, iv)
	_, err = c.Conn.Write(iv)
	return
}

func (c *conn) initDec() error {
	iv := c.buf[:c.cinfo.ivLen]
	n, err := c.Conn.Read(iv)
	if n != c.cinfo.ivLen || err != nil {
		return fmt.Errorf("bad iv(%d) %v, %v", n, iv[:n], err)
	}
	c.dec = c.cinfo.newDecStream(c.cinfo.block, iv)
	return nil
}
