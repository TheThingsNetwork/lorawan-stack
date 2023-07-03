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

package random_test

import (
	"testing"
	"time"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestPseudoRandom(t *testing.T) {
	a := assertions.New(t)
	a.So(Bytes(10), assertions.ShouldHaveLength, 10)

	a.So(Int63n(10), assertions.ShouldBeGreaterThanOrEqualTo, 0)
	a.So(Int63n(10), assertions.ShouldBeLessThan, 10)

	a.So(String(10), assertions.ShouldHaveLength, 10)
}

func BenchmarkBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Bytes(100)
	}
}

func BenchmarkIntn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int63n(100)
	}
}

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		String(100)
	}
}

func TestJitter(t *testing.T) {
	a := assertions.New(t)
	d := time.Duration(424242)
	p := 0.1
	for i := 0; i < 100; i++ {
		// Jitter of 10%
		t := Jitter(d, p)
		df := float64(d)
		a.So(t, should.BeBetweenOrEqual, df-df*p, df+df*p)
	}
}
