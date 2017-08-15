// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"encoding/json"
	"strings"
	"testing"

	. "github.com/smartystreets/assertions"
)

var levels = []Level{Debug, Info, Warn, Error, Fatal}

type o struct {
	Level Level `json:"level"`
}

func TestLevelParse(t *testing.T) {
	a := New(t)
	for _, level := range levels {
		str := level.String()

		{
			parsed, err := ParseLevel(str)
			a.So(err, ShouldBeNil)
			a.So(parsed, ShouldEqual, level)
		}

		{
			parsed, err := ParseLevel(strings.ToUpper(str))
			a.So(err, ShouldBeNil)
			a.So(parsed, ShouldEqual, level)
		}
	}
}

func TestLevelOrder(t *testing.T) {
	a := New(t)
	a.So(Info > Debug, ShouldBeTrue)
	a.So(Warn > Info, ShouldBeTrue)

	for _, level := range levels {
		a.So(level >= Debug, ShouldBeTrue)
		a.So(level <= Fatal, ShouldBeTrue)
		a.So(level < Debug, ShouldBeFalse)
		a.So(level > Fatal, ShouldBeFalse)
		a.So(level != invalid, ShouldBeTrue)
	}
}

func TestLevelJSONUnmarshal(t *testing.T) {
	for _, level := range levels {
		a := New(t)
		raw := []byte(`{ "level": "` + level.String() + `" }`)

		res := new(o)
		err := json.Unmarshal(raw, res)

		a.So(err, ShouldBeNil)
		a.So(res.Level, ShouldEqual, level)
	}
}

func TestLevelJSONMarshal(t *testing.T) {
	for _, level := range levels {
		a := New(t)

		raw := `{"level":"` + level.String() + `"}`

		res, err := json.Marshal(o{
			Level: level,
		})

		a.So(err, ShouldBeNil)
		a.So(string(res), ShouldEqual, raw)
	}
}
