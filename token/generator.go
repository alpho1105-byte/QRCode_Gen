package token

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type Generator interface {
	Generate(url string) (string, error)
}

type SHA256Generator struct {
	tokenLength int
}

// NewGenerator 接收 tokenLength，不再寫死
func NewGenerator(tokenLength int) Generator {
	return &SHA256Generator{tokenLength: tokenLength}
}

func (g *SHA256Generator) Generate(url string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	input := fmt.Sprintf("%s:%x", url, nonce)
	hash := sha256.Sum256([]byte(input))

	encoded := base62Encode(hash[:])

	if len(encoded) < g.tokenLength {
		return encoded, nil
	}
	return encoded[:g.tokenLength], nil
}

func base62Encode(data []byte) string {
	num := new(big.Int).SetBytes(data)
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := new(big.Int)

	var result []byte
	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, base62Chars[mod.Int64()])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}
