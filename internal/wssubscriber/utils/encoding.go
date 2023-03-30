package utils

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/btcsuite/btcutil/base58"
)

func Base64stringToHex(base64string string) string {
	data, _ := base64.StdEncoding.DecodeString(base64string)
	return "0x" + hex.EncodeToString(data)
}

func Base58stringToHex(base58string string) string {
	data := base58.Decode(base58string)
	return "0x" + hex.EncodeToString(data)
}

func Base64stringToBytes(base64string string) []byte {
	data, _ := base64.StdEncoding.DecodeString(base64string)
	return data
}

func Base64stringDecodeToString(base64string string) string {
	data, _ := base64.StdEncoding.DecodeString(base64string)
	return string(data)
}
