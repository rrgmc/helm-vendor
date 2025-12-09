package helm

import (
	"cmp"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"iter"
	"maps"
	"slices"
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

// MapSortedByKey returns an iterator for the given map that
// yields the key-value pairs in sorted order.
func MapSortedByKey[Map ~map[K]V, K cmp.Ordered, V any](m Map) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range slices.Sorted(maps.Keys(m)) {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}
