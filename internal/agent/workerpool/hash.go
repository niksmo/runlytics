package workerpool

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sync"
)

var hashPool = sync.Pool{}

func GetHashString(data []byte, key string) (string, error) {
	const op = "workerpool.GetHashString"
	h, ok := hashPool.Get().(hash.Hash)
	if !ok {
		h = hmac.New(sha256.New, []byte(key))
	} else {
		h.Reset()
	}
	defer hashPool.Put(h)
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
