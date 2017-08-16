// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package config

import (
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestUnmarshal(t *testing.T) {
	a := assertions.New(t)

	mgr := Initialize("test", defaults)
	a.So(mgr, should.NotBeNil)

	mgr.Parse()
	err := mgr.mergeConfig(strings.NewReader(`file-only: 10`))
	a.So(err, should.BeNil)

	var res map[string]interface{}
	err = mgr.Unmarshal(&res)
	a.So(err, should.BeNil)
	a.So(res, should.ContainKey, "file-only")
	a.So(res["file-only"], should.Resemble, 10)
}

func TestUnmarshalKey(t *testing.T) {
	a := assertions.New(t)

	mgr := Initialize("test", defaults)
	a.So(mgr, should.NotBeNil)

	mgr.Parse()
	err := mgr.mergeConfig(strings.NewReader(`file-only: 10`))
	a.So(err, should.BeNil)

	var res int
	err = mgr.UnmarshalKey("file-only", &res)
	a.So(err, should.BeNil)
	a.So(res, should.Resemble, 10)
}
