// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package pbkdf2

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestStringAlgorithm(t *testing.T) {
	a := New(t)

	a.So(Sha256.String(), ShouldEqual, "sha256")
	a.So(Sha512.String(), ShouldEqual, "sha512")
}

func TestParseAlgorithm(t *testing.T) {
	a := New(t)

	{
		alg, err := parseAlgorithm("sha256")
		a.So(err, ShouldBeNil)
		a.So(alg, ShouldResemble, Sha256)
	}
}

func TestParseBadAlgorithm(t *testing.T) {
	a := New(t)

	_, err := parseAlgorithm("bad")
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
