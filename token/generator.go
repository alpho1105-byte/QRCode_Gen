package token

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	tokenLength = 8 // 62^8 ≈ 2.18 × 10^14，遠超過 10 億筆需求
)

type Generator interface {
	Generate(url string) (string, error)
}

type SHA256Generator struct{}

func NewGenerator() Generator {
	return &SHA256Generator{}
}

func (g *SHA256Generator) Generate(url string) (string, error) {
	//todo
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// URL + nonce 組合後做 SHA-256
	input := fmt.Sprintf("%s:%x", url, nonce)
	hash := sha256.Sum256([]byte(input))

	// Step 3: Base62 編碼
	encoded := base62Encode(hash[:])

	if len(encoded) < tokenLength {
		return encoded, nil
	}
	return encoded[:tokenLength], nil // ← 取前 8 個字元
}

func base62Encode(data []byte) string {
	num := new(big.Int).SetBytes(data)
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := new(big.Int)

	var result []byte
	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, []byte(base62Chars)[mod.Int64()])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

/*
如果只對 URL 做 SHA-256，同一個 URL 永遠得到同一個 hash。這樣兩個使用者都建立 `https://google.com` 就會碰撞。加了隨機 nonce 之後：
```
"https://google.com:a3f2b1..."  → hash A
"https://google.com:7c8d9e..."  → hash B  （不同 nonce → 不同結果）
*/
