// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package filter

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"

	yaml "gopkg.in/yaml.v1"
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
