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

package log_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/config"
	. "go.thethings.network/lorawan-stack/pkg/log"
)

var _ config.Configurable = func(v Level) *Level { return &v }(0)

var levels = []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}

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

	a.So(InfoLevel > DebugLevel, should.BeTrue)
	a.So(WarnLevel > InfoLevel, should.BeTrue)

	for _, level := range levels {
		a.So(level, should.BeGreaterThanOrEqualTo, DebugLevel)
		a.So(level, should.BeLessThanOrEqualTo, FatalLevel)
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
