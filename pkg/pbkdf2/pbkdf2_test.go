// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package pbkdf2

import (
	"testing"

	. "github.com/smartystreets/assertions"
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
