package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type SecurityProvider interface {
	EncryptData(dataIv []byte, originData []byte) []byte
	DecryptData(dataIv []byte, encryptedData []byte) []byte
	RandomBytes(size int) []byte
	BlockSize() int

	EncryptUrlData(urlData string, userID uint64) string
}

type SimpleCryptoProvider struct {
	blockCipher cipher.Block
	dataIv      []byte
}

func NewSimpleCryptoProvider(dataKey, dataIv string) SecurityProvider {
	key, err := base64.StdEncoding.DecodeString(dataKey)
	if err != nil {
		panic(errors.WithMessage(err, "new simple crypto provider failed, decode dataKey error"))
	}
	iv, err := base64.StdEncoding.DecodeString(dataIv)
	if err != nil {
		panic(errors.WithMessage(err, "new simple crypto provider failed, decode dataIv error"))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(errors.WithMessage(err, "new simple crypto provider failed"))
	}
	return &SimpleCryptoProvider{
		blockCipher: block,
		dataIv:      iv,
	}
}

// EncryptData 加密数据
func (s *SimpleCryptoProvider) EncryptData(dataIv []byte, originData []byte) []byte {
	input := []byte(originData)
	// 填充数据
	paddedData := padPKCS7(input, aes.BlockSize)
	// 使用 CBC 模式加密
	mode := cipher.NewCBCEncrypter(s.blockCipher, dataIv)
	encryptedData := make([]byte, len(paddedData))
	mode.CryptBlocks(encryptedData, paddedData)
	return encryptedData
}

func (s *SimpleCryptoProvider) DecryptData(dataIv []byte, encryptedData []byte) []byte {
	// 创建 CBC 模式解密器
	mode := cipher.NewCBCDecrypter(s.blockCipher, dataIv)
	// 解密数据
	decryptedData := make([]byte, len(encryptedData))
	mode.CryptBlocks(decryptedData, encryptedData)
	// 去除填充
	decryptedData = unpadPKCS7(decryptedData)
	return decryptedData
}

func (s *SimpleCryptoProvider) RandomBytes(size int) []byte {
	buffer := make([]byte, size)
	_, err := rand.Read(buffer)
	if err != nil {
		panic(err)
	}
	return buffer
}

func (s *SimpleCryptoProvider) BlockSize() int {
	return s.blockCipher.BlockSize()
}

func (s *SimpleCryptoProvider) EncryptUrlData(urlData string, userID uint64) string {
	input := fmt.Sprintf("%s#%d#%v#sardine", urlData, userID, time.Now().Nanosecond())
	return base64.URLEncoding.EncodeToString(s.EncryptData(s.dataIv, []byte(input)))
}

func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// 去除 PKCS7 填充
func unpadPKCS7(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}
