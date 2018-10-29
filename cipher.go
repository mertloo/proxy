package ssocks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"errors"
)

var ErrNotSupportMeth = errors.New("not support cipher method")

type cipherInfo struct {
	ivLen        int
	keyLen       int
	key          []byte
	newCipher    func(key []byte) (block cipher.Block, err error)
	newEncStream func(block cipher.Block, iv []byte) (stream cipher.Stream)
	newDecStream func(block cipher.Block, iv []byte) (stream cipher.Stream)
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
	cinfo, ok := cipherInfoMap[method]
	if !ok {
		err = ErrNotSupportMeth
		return
	}
	cinfo.key = evpBytesToKey(password, cinfo.keyLen)
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

/*
func NewEncrypter(info *cipherInfo, iv []byte) (encrypter Encrypter, err error) {
	block, err := info.newCipher(info.key)
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

func NewDecrypter(info *cipherInfo, iv []byte) (decrypter Decrypter, err error) {
	block, err := info.newCipher(info.key)
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
*/
