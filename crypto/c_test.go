package crypto

import (
	"crypto/rand"
	"testing"

	"encoding/base64"
	"fmt"
)

var SimpleSignKeyProvider SecurityProvider

func InitSimpleCryptoProvider(dataKey, dataIv string) {
	SimpleSignKeyProvider = NewSimpleCryptoProvider(dataKey, dataIv)
}

func Random(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("Error generating random bytes:", err)
		return nil, err
	}
	return randomBytes, nil
}

func createData() (encodeB32, encodeB16 string) {
	b32, err := Random(32)
	if err != nil {
		panic(err)
	}
	b16, err := Random(16)
	if err != nil {
		panic(err)
	}
	encodeB32 = base64.StdEncoding.EncodeToString(b32)
	encodeB16 = base64.StdEncoding.EncodeToString(b16)
	return encodeB32, encodeB16
}

func TestGenerateCryptoStr(t *testing.T) {
	encodeB32, encodeB16 := createData()
	InitSimpleCryptoProvider(encodeB32, encodeB16)
	randomB32 := SimpleSignKeyProvider.RandomBytes(32)
	randomB16 := SimpleSignKeyProvider.RandomBytes(16)
	dataKey := base64.StdEncoding.EncodeToString(randomB32)
	dataIv := base64.StdEncoding.EncodeToString(randomB16)
	fmt.Println("dataKey = ", dataKey)
	fmt.Println("dataIv = ", dataIv)
}
