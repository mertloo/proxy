package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

type SSocks struct{}

func (ss *SSocks) NewDConn(conn net.Conn, password string, info *cipherInfo) (dConn *DConn, err error) {
	buf := make([]byte, info.ivLen)
	n, err := conn.Read(buf)
	if n != info.ivLen || err != nil {
		err = fmt.Errorf("read iv error (read: %v, err: %v)", buf[:n], err)
		return
	}
	key, iv := EVPBytesToKey(password, info.keyLen), buf[:n]
	decrypter, err := NewDecrypter(info, key, iv)
	if err != nil {
		return
	}
	dConn = &DConn{conn, decrypter}
	return
}

func (ss *SSocks) NewEConn(conn net.Conn, password string, info *cipherInfo) (eConn *EConn, err error) {
	buf := make([]byte, info.ivLen)
	_, err = io.ReadFull(rand.Reader, buf)
	if err != nil {
		return
	}
	n, err := conn.Write(buf)
	if n != info.ivLen || err != nil {
		err = fmt.Errorf("write iv error (write: %v, err: %v)", buf[:n], err)
		return
	}
	key, iv := EVPBytesToKey(password, info.keyLen), buf[:n]
	encrypter, err := NewEncrypter(info, key, iv)
	if err != nil {
		return
	}
	eConn = &EConn{conn, encrypter}
	return
}

func (ss *SSocks) Connect(conn net.Conn, dialer Dialer) (dstConn net.Conn, err error) {
	if _, ok := conn.(*DConn); !ok {
		err = fmt.Errorf("conn not *DConn")
		return
	}
	addr, err := ReadAddr(conn)
	if err != nil {
		return
	}
	dstConn, err = dialer.Dial(addr)
	return
}
