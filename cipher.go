package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"fmt"
	"io"
)

func md5sum(d []byte) []byte {
	h := md5.New()
	h.Write(d)
	return h.Sum(nil)
}

func evpBytesToKey(password string, keyLen int) (key []byte) {
	const md5Len = 16

	cnt := (keyLen-1)/md5Len + 1
	m := make([]byte, cnt*md5Len)
	copy(m, md5sum([]byte(password)))

	// Repeatedly call md5 until bytes generated is enough.
	// Each call to md5 uses data: prev md5 sum + password.
	d := make([]byte, md5Len+len(password))
	start := 0
	for i := 1; i < cnt; i++ {
		start += md5Len
		copy(d, m[start-md5Len:start])
		copy(d[md5Len:], password)
		copy(m[start:], md5sum(d))
	}
	return m[:keyLen]
}

type cryptInfo struct {
	newCipherFunc func(iv []byte, password string) (*Cipher, error)
	ivLen         int
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
	return info.newCipherFunc(buf[:n], password)
}

func newAES256CFB(iv []byte, password string) (cf *Cipher, err error) {
	key := evpBytesToKey(password, 32)
	block, err := aes.NewCipher(key)
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
