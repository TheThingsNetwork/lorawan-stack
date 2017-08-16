// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var levels = []Level{Debug, Info, Warn, Error, Fatal}

type o struct {
	Level Level `json:"level"`
}

func TestLevelParse(t *testing.T) {
	a := assertions.New(t)

	for _, level := range levels {
		str := level.String()

		{
			parsed, err := ParseLevel(str)
			a.So(err, should.BeNil)
			a.So(parsed, should.Equal, level)
		}

		{
			parsed, err := ParseLevel(strings.ToUpper(str))
			a.So(err, should.BeNil)
			a.So(parsed, should.Equal, level)
		}
	}
}

func TestLevelOrder(t *testing.T) {
	a := assertions.New(t)

	a.So(Info > Debug, should.BeTrue)
	a.So(Warn > Info, should.BeTrue)

	for _, level := range levels {
		a.So(level >= Debug, should.BeTrue)
		a.So(level <= Fatal, should.BeTrue)
		a.So(level < Debug, should.BeFalse)
		a.So(level > Fatal, should.BeFalse)
		a.So(level != invalid, should.BeTrue)
	}
}

func TestLevelJSONUnmarshal(t *testing.T) {
	a := assertions.New(t)

	for _, level := range levels {
		raw := []byte(`{ "level": "` + level.String() + `" }`)

		res := new(o)
		err := json.Unmarshal(raw, res)

		a.So(err, should.BeNil)
		a.So(res.Level, should.Equal, level)
	}
}

func TestLevelJSONMarshal(t *testing.T) {
	a := assertions.New(t)

	for _, level := range levels {

		raw := `{"level":"` + level.String() + `"}`

		res, err := json.Marshal(o{
			Level: level,
		})

		a.So(err, should.BeNil)
		a.So(string(res), should.Equal, raw)
	}
}
