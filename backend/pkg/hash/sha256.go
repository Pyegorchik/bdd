package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func (h *hashManager) HashSha256(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (h *hashManager) TriplePassHash(nakedPass string) string {
	first := sha256.Sum256([]byte(nakedPass))
	second := sha256.Sum256(first[:])
	return h.HashSha256(hex.EncodeToString(second[:]))
}