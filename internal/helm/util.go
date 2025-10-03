package helm

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

func randomName() string {
	buf := make([]byte, 20)
	_, _ = rand.Read(buf)
	return strings.ReplaceAll(base64.StdEncoding.EncodeToString(buf), "/", "-")
}

func JoinHTTPPaths(baseURL, paths string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(strings.TrimSpace(baseURL), "/"), paths)
}
