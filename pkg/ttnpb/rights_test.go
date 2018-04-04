// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRightStringer(t *testing.T) {
	a := assertions.New(t)
	a.So(RIGHT_APPLICATION_DELETE.String(), should.Equal, "RIGHT_APPLICATION_DELETE")
	a.So(Right(1234).String(), should.Equal, "1234")
	a.So(RIGHT_APPLICATION_DELETE.String(), should.Equal, Right_name[int32(RIGHT_APPLICATION_DELETE)])
	a.So(Right(1234).String(), should.Equal, "1234")
}

func TestRightText(t *testing.T) {
	a := assertions.New(t)

	text, err := RIGHT_APPLICATION_DELETE.MarshalText()
	a.So(err, should.BeNil)
	a.So(string(text), should.Equal, "RIGHT_APPLICATION_DELETE")

	var right Right
	err = (&right).UnmarshalText([]byte("RIGHT_APPLICATION_DELETE"))
	a.So(err, should.BeNil)
	a.So(right, should.Equal, RIGHT_APPLICATION_DELETE)

	err = right.UnmarshalText([]byte("foo"))
	a.So(err, should.NotBeNil)
}

func TestRightJSON(t *testing.T) {
	a := assertions.New(t)

	b, err := json.Marshal(RIGHT_APPLICATION_DELETE)
	a.So(err, should.BeNil)
	a.So(b, should.Resemble, []byte(`"RIGHT_APPLICATION_DELETE"`))

	var right Right
	err = json.Unmarshal([]byte(`"RIGHT_APPLICATION_DELETE"`), &right)
	a.So(err, should.BeNil)
	a.So(right, should.Equal, RIGHT_APPLICATION_DELETE)

	err = json.Unmarshal([]byte(`"foo"`), right)
	a.So(err, should.NotBeNil)
}
