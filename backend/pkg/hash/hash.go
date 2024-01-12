package hash

type HashManager interface {
	HashSha256(s string) string
	TriplePassHash(nakedPass string) string
}

type hashManager struct {
}

func NewHashManager() HashManager {
	return &hashManager{}
}