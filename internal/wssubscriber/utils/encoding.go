package utils

import (
  "encoding/hex"
  "encoding/base64"
)

func Base64stringToHex(base64string string) string {
  data, _ := base64.StdEncoding.DecodeString(base64string)
  return "0x" + hex.EncodeToString(data)
}
