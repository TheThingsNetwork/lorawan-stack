// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRightNormalization(t *testing.T) {
	a := assertions.New(t)
	a.So(normalizeRight("APPLICATION_DELETE"), should.Equal, "application:delete")
	a.So(denormalizeRight("application:delete"), should.Equal, "APPLICATION_DELETE")
}

func TestRightStringer(t *testing.T) {
	a := assertions.New(t)
	a.So(RightApplicationDelete.String(), should.Equal, "application:delete")
	a.So(Right(1234).String(), should.Equal, "1234")
}

func TestRightText(t *testing.T) {
	a := assertions.New(t)

	text, err := RightApplicationDelete.MarshalText()
	a.So(err, should.BeNil)
	a.So(string(text), should.Equal, "application:delete")

	var right Right
	err = (&right).UnmarshalText([]byte("application:delete"))
	a.So(err, should.BeNil)
	a.So(right, should.Equal, RightApplicationDelete)

	err = right.UnmarshalText([]byte("foo"))
	a.So(err, should.NotBeNil)
}

func TestRightJSON(t *testing.T) {
	a := assertions.New(t)

	b, err := json.Marshal(RightApplicationDelete)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte(`"application:delete"`))

	var right Right
	err = json.Unmarshal([]byte(`"application:delete"`), &right)
	a.So(err, should.BeNil)
	a.So(right, should.Equal, RightApplicationDelete)

	err = json.Unmarshal([]byte(`"foo"`), right)
	a.So(err, should.NotBeNil)
}
