package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func HashName(bytes []byte) string {
	var hasher = sha256.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func Base64Decode(encoded []byte) []byte{
	var decoded = make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	if _, err := base64.StdEncoding.Decode(decoded, encoded); err != nil {
		return nil 
	} else {
		return decoded
	}
}
