// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRightStringer(t *testing.T) {
	a := assertions.New(t)
	a.So(RIGHT_APPLICATION_DELETE.TextString(), should.Equal, "application:delete")
	a.So(Right(1234).TextString(), should.Equal, "1234")
	a.So(RIGHT_APPLICATION_DELETE.String(), should.Equal, Right_name[int32(RIGHT_APPLICATION_DELETE)])
	a.So(Right(1234).String(), should.Equal, "1234")
}

func TestRightText(t *testing.T) {
	a := assertions.New(t)

	text, err := RIGHT_APPLICATION_DELETE.MarshalText()
	a.So(err, should.BeNil)
	a.So(string(text), should.Equal, "application:delete")

	var right Right
	err = (&right).UnmarshalText([]byte("application:delete"))
	a.So(err, should.BeNil)
	a.So(right, should.Equal, RIGHT_APPLICATION_DELETE)

	err = right.UnmarshalText([]byte("foo"))
	a.So(err, should.NotBeNil)
}

func TestRightJSON(t *testing.T) {
	a := assertions.New(t)

	b, err := json.Marshal(RIGHT_APPLICATION_DELETE)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte(`"application:delete"`))

	var right Right
	err = json.Unmarshal([]byte(`"application:delete"`), &right)
	a.So(err, should.BeNil)
	a.So(right, should.Equal, RIGHT_APPLICATION_DELETE)

	err = json.Unmarshal([]byte(`"foo"`), right)
	a.So(err, should.NotBeNil)
}
