package block

import (
	"encoding/hex"
	"strings"
)

func decodeExtraData(data string) (string, bool) {
	hexStr := strings.TrimPrefix(data, "0x")
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", false
	}
	return string(b), true
}
