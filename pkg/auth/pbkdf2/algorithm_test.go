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

package pbkdf2_test

import (
	"testing"

	. "github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
)

func TestStringAlgorithm(t *testing.T) {
	a := New(t)

	a.So(Sha256.String(), ShouldEqual, "sha256")
	a.So(Sha512.String(), ShouldEqual, "sha512")
}

func TestParseAlgorithm(t *testing.T) {
	a := New(t)

	{
		alg, err := ParseAlgorithm("sha256")
		a.So(err, ShouldBeNil)
		a.So(alg, ShouldResemble, Sha256)
	}
}

func TestParseBadAlgorithm(t *testing.T) {
	a := New(t)

	_, err := ParseAlgorithm("bad")
	a.So(err, ShouldNotBeNil)
}

func TestNil(t *testing.T) {
	a := New(t)

	{
		var alg Algorithm
		h := alg.Hash()
		a.So(h, ShouldBeNil)
	}

	{
		var alg *Algorithm
		h := alg.Hash()
		a.So(h, ShouldBeNil)
	}
}
