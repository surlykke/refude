package encode
// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/surlykke/refude/internal/lib/xdg"
)

func HashName(bytes []byte) string {
	var hasher = sha256.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func Base64Decode(encoded []byte) ([]byte, error) {
	var decoded = make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	if _, err := base64.StdEncoding.Decode(decoded, encoded); err != nil {
		return nil, err
	} else {
		return decoded, nil
	}
}

func WriteCasheFile(bytes []byte, fileType string) (string, error) {
	var osPath = xdg.RefudeCacheDir + "/" + HashName(bytes) + "." + fileType
	if err := os.WriteFile(osPath, bytes, 0600); err != nil {
		return "", err
	} else {
		return osPath, nil
	}
}
