// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package random

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
)

// Interface for random
type Interface interface {
	Intn(n int) int
	String(n int) string
	Bytes(n int) []byte
	FillBytes(p []byte)
}

// TTNRandom is used as a wrapper around crypto/rand
type TTNRandom struct {
	Source io.Reader
}

// New returns a new Random, in most cases you can also just use the global funcs
func New() Interface {
	return &TTNRandom{
		Source: rand.Reader,
	}
}

var global = New()

// Intn returns a random number in the range [0,n)
func Intn(n int) int { return global.Intn(n) }
func (r *TTNRandom) Intn(n int) int {
	i, _ := rand.Int(r.Source, big.NewInt(int64(n)))
	return int(i.Int64())
}

// Bytes generates a random byte slice of length n
func Bytes(n int) []byte { return global.Bytes(n) }
func (r *TTNRandom) Bytes(n int) []byte {
	p := make([]byte, n)
	r.FillBytes(p)
	return p
}

// FillBytes fills the byte slice with random bytes. It does not use an intermediate buffer
func FillBytes(p []byte) { global.FillBytes(p) }
func (r *TTNRandom) FillBytes(p []byte) {
	_, err := r.Source.Read(p)
	if err != nil {
		panic(fmt.Errorf("random.Bytes: %s", err))
	}
}

// String returns a random string of length n, it uses the characters of base64.URLEncoding
func String(n int) string { return global.String(n) }
func (r *TTNRandom) String(n int) string {
	b := r.Bytes(n * 6 / 8)
	return base64.RawURLEncoding.EncodeToString(b)
}
