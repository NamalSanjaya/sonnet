package bicrypt

import (
	"encoding/hex"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type Interface interface {
	Encrypt(plainText string) string
	Decrypt(cipherText  string) (string, error)
	HexToUft8(hexStr string) (string, error)
}

type client struct {
	aead cipher.AEAD
	nonce []byte
}

type Configs struct {
	AesKey string         // 16 characters secret key
	NonceKey string       // 12 characters long nonce
} 

var _ Interface = (*client)(nil)

func New(configs *Configs) Interface {
	cipherBlock, err := aes.NewCipher([]byte(configs.AesKey))
    if err != nil {
        panic(err)
    }

	aesGcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		panic(err)
	}
	return &client{
		aead: aesGcm,
		nonce: []byte(configs.NonceKey),
	}
}

// encrpt and encode to hexdecimal
func (b *client) Encrypt(plainText string) string{
	return fmt.Sprintf("%x", b.aead.Seal(nil, b.nonce, []byte(plainText), nil))
}

// decrpt and decode to utf-8 text
func (b *client) Decrypt(cipherText string) (string, error){
	decrpt, err := b.aead.Open(nil, b.nonce, []byte(cipherText), nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(decrpt), nil
}

// hex encoding to utf8 text
func (b *client) HexToUft8(hexStr string) (string, error) {
	src := []byte(hexStr)
	dst := make([]byte, hex.DecodedLen(len(src)))
	if _, err := hex.Decode(dst, src); err != nil {
		return "", err
	}
	return string(dst), nil
}
