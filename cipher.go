package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
)

type cryptInfo struct {
	newFunc func(iv []byte, password string) (*Cipher, error)
	ivLen   int
}

var (
	cryptMap = map[string]cryptInfo{
		"aes256cfb": {newAES256CFB, aes.BlockSize},
	}
)

func getCryptInfo(cryptMeth string) (info cryptInfo, err error) {
	info, ok := cryptMap[cryptMeth]
	if !ok {
		err = fmt.Errorf("not support crypt meth %s", cryptMeth)
		return
	}
	return
}

type Cipher struct {
	encryptStream cipher.Stream
	decryptStream cipher.Stream
}

func newCipher(cryptMeth, password string, ivReader io.Reader) (cipher *Cipher, err error) {
	info, ok := cryptMap[cryptMeth]
	if !ok {
		return nil, fmt.Errorf("not support crypt meth %s", cryptMeth)
	}
	buf := make([]byte, info.ivLen)
	n, err := ivReader.Read(buf)
	if n == 0 || err != nil {
		return nil, fmt.Errorf("read iv err")
	}
	return info.newFunc(buf[:n], password)
}

func newAES256CFB(iv []byte, password string) (cf *Cipher, err error) {
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return
	}
	cf = new(Cipher)
	cf.encryptStream = cipher.NewCFBEncrypter(block, iv)
	cf.decryptStream = cipher.NewCFBDecrypter(block, iv)
	return
}

func (cf *Cipher) Encrypt(dst, src []byte) {
	cf.encryptStream.XORKeyStream(dst, src)
}

func (cf *Cipher) Decrypt(dst, src []byte) {
	cf.decryptStream.XORKeyStream(dst, src)
}
