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

package pbkdf2

import (
	"testing"

	. "github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestName(t *testing.T) {
	a := New(t)
	h := &PBKDF2{}
	a.So(h.Name(), ShouldEqual, "PBKDF2")
}

func TestHash(t *testing.T) {
	a := New(t)
	hashes := []Algorithm{Sha512, Sha256}
	for _, hash := range hashes {

		h := &PBKDF2{
			Iterations: 10000,
			KeyLength:  512,
			Algorithm:  hash,
			SaltLength: 64,
		}

		plain := "secret"

		hashed, err := h.Hash(plain)
		a.So(err, ShouldBeNil)

		// should validate against plain
		{
			ok, err := h.Validate(hashed, plain)
			a.So(err, ShouldBeNil)
			a.So(ok, ShouldBeTrue)
		}

		// should not validate against wrong plain
		{
			other, err := h.Hash("othersecret")
			a.So(err, should.BeNil)
			ok, err := h.Validate(other, plain)
			a.So(err, ShouldBeNil)
			a.So(ok, ShouldBeFalse)
		}

		// should not parse a bad format
		{
			ok, err := h.Validate("badformat", "somethingelse")
			a.So(err, ShouldNotBeNil)
			a.So(ok, ShouldBeFalse)
		}
	}
}

func TestHashZeroSalt(t *testing.T) {
	a := New(t)

	h := &PBKDF2{
		Iterations: 10000,
		KeyLength:  512,
		Algorithm:  Sha512,
		SaltLength: 0,
	}

	plain := "secret"

	_, err := h.Hash(plain)
	a.So(err, ShouldNotBeNil)
}

func TestBadHash(t *testing.T) {
	a := New(t)

	h := &PBKDF2{
		Iterations: 10000,
		KeyLength:  512,
		Algorithm:  Sha512,
		SaltLength: 36,
	}

	// bad hash, the base64 is wrong
	hashed := "PBKDF2$sha512$10000$08ThlOIywy64D3C7m9SPQRybJgAZOgGk49j4yk8HQr10XJar$foo=="
	plain := "foo"

	ok, err := h.Validate(hashed, plain)
	a.So(err, ShouldNotBeNil)
	a.So(ok, ShouldBeFalse)
}

func TestBadIter(t *testing.T) {
	a := New(t)

	h := &PBKDF2{
		Iterations: 10000,
		KeyLength:  512,
		Algorithm:  Sha512,
		SaltLength: 36,
	}

	// bad hash, the salt is bad
	plain := "foo"
	hashed := "PBKDF2$sha512$bad$08ThlOIywy64D3C7m9SPQRybJgAZOgGk49j4yk8HQr10XJar$foo=="

	ok, err := h.Validate(hashed, plain)
	a.So(err, ShouldNotBeNil)
	a.So(ok, ShouldBeFalse)
}

func TestBadAlgorithm(t *testing.T) {
	a := New(t)

	h := &PBKDF2{
		Iterations: 10000,
		KeyLength:  512,
		Algorithm:  Sha512,
		SaltLength: 36,
	}

	// bad hash, the base64 is bad
	plain := "foo"
	hashed := "PBKDF2$bad$1000$08ThlOIywy64D3C7m9SPQRybJgAZOgGk49j4yk8HQr10XJar$foo=="

	ok, err := h.Validate(hashed, plain)
	a.So(err, ShouldNotBeNil)
	a.So(ok, ShouldBeFalse)
}

func TestInvalidSaltLength(t *testing.T) {
	a := New(t)

	h := &PBKDF2{
		SaltLength: 0,
	}

	_, err := h.Hash("foo")
	a.So(err, ShouldNotBeNil)
}
