// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package random implements goroutine-safe utilities on top of a secure random source.
package random

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"time"
)

// Int63n returns a random number in the range [0,n).
func Int63n(n int64) int64 {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return i.Int64()
}

// Bytes generates a random byte slice of length n.
func Bytes(n int) []byte {
	p := make([]byte, n)
	_, err := rand.Read(p)
	if err != nil {
		panic(err)
	}
	return p
}

// String returns a random string of length n, it uses the characters of base64.URLEncoding.
func String(n int) string {
	b := Bytes(n * 6 / 8)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Jitter returns a random number around d where p is the maximum percentage of change applied to d.
// With d=100 and p=0.1, the duration returned will be in [90,110].
func Jitter(d time.Duration, p float64) time.Duration {
	df := float64(d)
	v := time.Duration(Int63n(int64(df*p*2)) - int64(df*p))
	return d + v
}

// CanJitter checks if the provided duration `d` can be used with the Jitter function with the provided
// percentage p.
func CanJitter(d time.Duration, p float64) bool {
	return int64(float64(d)*p*2) > 0
}
