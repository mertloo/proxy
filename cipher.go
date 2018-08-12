package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"fmt"
)

type newCipherFunc func(key []byte) (block cipher.Block, err error)
type newStreamFunc func(block cipher.Block, iv []byte) (stream cipher.Stream)

type cipherInfo struct {
	ivLen        int
	keyLen       int
	newCipher    newCipherFunc
	newEncStream newStreamFunc
	newDecStream newStreamFunc
}

type Encrypter interface {
	Encrypt(dst, src []byte)
}

type Decrypter interface {
	Decrypt(dst, src []byte)
}

type StreamEncrypter struct {
	cipher.Stream
}

func (enc *StreamEncrypter) Encrypt(dst, src []byte) {
	enc.XORKeyStream(dst, src)
	return
}

type StreamDecrypter struct {
	cipher.Stream
}

func (dec *StreamDecrypter) Decrypt(dst, src []byte) {
	dec.XORKeyStream(dst, src)
	return
}

var cipherInfoMap = map[string]*cipherInfo{
	"aes256cfb": &cipherInfo{aes.BlockSize, 32, aes.NewCipher, cipher.NewCFBEncrypter, cipher.NewCFBDecrypter},
}

func GetCipherInfo(cipherName string) (info *cipherInfo, err error) {
	info, ok := cipherInfoMap[cipherName]
	if !ok {
		err = fmt.Errorf("no support cipher %s", cipherName)
	}
	return
}

func EVPBytesToKey(password string, keyLen int) (key []byte) {
	const md5Len = 16

	cnt := (keyLen-1)/md5Len + 1
	m := make([]byte, cnt*md5Len)
	s := md5.Sum([]byte(password))
	copy(m, s[:])

	// Repeatedly call md5 until bytes generated is enough.
	// Each call to md5 uses data: prev md5 sum + password.
	d := make([]byte, md5Len+len(password))
	start := 0
	for i := 1; i < cnt; i++ {
		start += md5Len
		copy(d, m[start-md5Len:start])
		copy(d[md5Len:], password)
		s = md5.Sum(d)
		copy(m[start:], s[:])
	}
	return m[:keyLen]
}

func NewEncrypter(info *cipherInfo, key, iv []byte) (encrypter Encrypter, err error) {
	block, err := info.newCipher(key)
	if err != nil {
		return
	}
	if info.newEncStream == nil {
		encrypter = block
		return
	}
	stream := info.newEncStream(block, iv)
	encrypter = &StreamEncrypter{stream}
	return
}

func NewDecrypter(info *cipherInfo, key, iv []byte) (decrypter Decrypter, err error) {
	block, err := info.newCipher(key)
	if err != nil {
		return
	}
	if info.newDecStream == nil {
		decrypter = block
		return
	}
	stream := info.newDecStream(block, iv)
	decrypter = &StreamDecrypter{stream}
	return
}
