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

package filter

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	yaml "gopkg.in/yaml.v2"
)

func TestBuild(t *testing.T) {
	a := assertions.New(t)

	raw := []byte(`
and:
  - match:
      field: foo
      value: "x"
  - or:
    - match:
        field: bar
        value: "y"
    - match:
        field: baz
        value: "z"
`)

	var m map[string]interface{}
	err := yaml.Unmarshal(raw, &m)
	a.So(err, should.BeNil)

	_, err = Build(m)
	a.So(err, should.BeNil)

	_, err = Build(m["and"])
	a.So(err, should.BeNil)
}
