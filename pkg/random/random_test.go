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

package random

import (
	"testing"

	"github.com/smartystreets/assertions"
)

func TestPseudoRandom(t *testing.T) {
	a := assertions.New(t)
	r := New()

	a.So(Bytes(10), assertions.ShouldHaveLength, 10)
	a.So(r.Bytes(10), assertions.ShouldHaveLength, 10)

	a.So(Intn(10), assertions.ShouldBeGreaterThanOrEqualTo, 0)
	a.So(r.Intn(10), assertions.ShouldBeGreaterThanOrEqualTo, 0)
	a.So(Intn(10), assertions.ShouldBeLessThan, 10)
	a.So(r.Intn(10), assertions.ShouldBeLessThan, 10)

	a.So(String(10), assertions.ShouldHaveLength, 10)
	a.So(r.String(10), assertions.ShouldHaveLength, 10)

	p := make([]byte, 100)
	Read(p)
	a.So(p, assertions.ShouldNotResemble, make([]byte, 100))

	q := make([]byte, 100)
	r.Read(q)
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

func BenchmarkRead(b *testing.B) {
	p := make([]byte, 100)
	for i := 0; i < b.N; i++ {
		Read(p)
	}
}
