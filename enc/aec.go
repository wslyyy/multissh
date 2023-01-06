package enc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func AesEncEncode(origData, key []byte) (string, error) {
	si, err := AesEncrypt(origData, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(si), nil
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func AesDecEncode(encodeStr string, key []byte) ([]byte, error) {
	crypted, err := base64.StdEncoding.DecodeString(encodeStr)
	if err != nil {
		return nil, err
	}
	return AesDecrypt(crypted, key)
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = ZeroUnPadding(origData)
	return origData, nil
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	for length > 0 {
		unpadding := int(origData[length-1])
		if unpadding == 0 && length > 0 {
			length = length - 1
		} else {
			break
		}
	}
	if length == 0 {
		return origData[:1]
	} else {
		return origData[:length]
	}
}
