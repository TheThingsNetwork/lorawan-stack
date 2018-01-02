// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package pseudorandom

import (
	"testing"

	"github.com/smartystreets/assertions"
)

func TestPseudoRandom(t *testing.T) {
	a := assertions.New(t)
	r := New(1)

	a.So(Bytes(10), assertions.ShouldHaveLength, 10)
	a.So(r.Bytes(10), assertions.ShouldHaveLength, 10)

	a.So(Intn(10), assertions.ShouldBeGreaterThanOrEqualTo, 0)
	a.So(r.Intn(10), assertions.ShouldBeGreaterThanOrEqualTo, 0)
	a.So(Intn(10), assertions.ShouldBeLessThan, 10)
	a.So(r.Intn(10), assertions.ShouldBeLessThan, 10)

	a.So(String(10), assertions.ShouldHaveLength, 10)
	a.So(r.String(10), assertions.ShouldHaveLength, 10)

	p := make([]byte, 100)
	FillBytes(p)
	a.So(p, assertions.ShouldNotResemble, make([]byte, 100))

	q := make([]byte, 100)
	r.FillBytes(q)
	a.So(q, assertions.ShouldNotResemble, make([]byte, 100))
}

func BenchmarkBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Bytes(100)
	}
}

func BenchmarkIntn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intn(100)
	}
}

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		String(100)
	}
}

func BenchmarkFillBytes(b *testing.B) {
	p := make([]byte, 100)
	for i := 0; i < b.N; i++ {
		FillBytes(p)
	}
}
