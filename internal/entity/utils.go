package entity

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
)

func generateId(path string) string {
	cleanPath := strings.TrimRight(path, string(os.PathSeparator))
	h := sha256.New()
	h.Write([]byte(cleanPath))
	id := fmt.Sprintf("%x", h.Sum(nil))
	return string(id[:12])
}
