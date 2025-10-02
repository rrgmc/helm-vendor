package helm

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func randomName() string {
	buf := make([]byte, 20)
	_, _ = rand.Read(buf)
	return strings.ReplaceAll(base64.StdEncoding.EncodeToString(buf), "/", "-")
}
