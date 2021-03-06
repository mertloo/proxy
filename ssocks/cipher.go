package ssocks

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
	key          []byte
	block        cipher.Block
	newCipher    newCipherFunc
	newEncStream newStreamFunc
	newDecStream newStreamFunc
}

var cipherInfoMap = map[string]*cipherInfo{
	"aes256cfb": &cipherInfo{
		ivLen:        aes.BlockSize,
		keyLen:       32,
		key:          nil,
		newCipher:    aes.NewCipher,
		newEncStream: cipher.NewCFBEncrypter,
		newDecStream: cipher.NewCFBDecrypter,
	},
}

func getCipherInfo(method, password string) (cinfo *cipherInfo, err error) {
	cinfo = cipherInfoMap[method]
	if cinfo == nil {
		err = fmt.Errorf("not supported cipher: %s", method)
		return
	}
	cinfo.key = evpBytesToKey(password, cinfo.keyLen)
	cinfo.block, err = cinfo.newCipher(cinfo.key)
	if err != nil {
		cinfo = nil
	}
	return
}

func evpBytesToKey(password string, keyLen int) (key []byte) {
	const md5Len = 16

	cnt := (keyLen-1)/md5Len + 1
	m := make([]byte, cnt*md5Len)
	s := md5.Sum([]byte(password))
	copy(m, s[:])

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
