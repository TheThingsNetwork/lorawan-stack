// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package config

import (
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type singleFileConfig struct {
	ConfigPath string            `name:"config"`
	Foo        string            `name:"foo"`
	Bar        map[string]string `name:"bar"`
}

func TestReadSingleConfig(t *testing.T) {
	a := assertions.New(t)

	_, filename, _, _ := runtime.Caller(0)

	first := path.Join(filepath.Dir(filename), "first.yml")
	second := path.Join(filepath.Dir(filename), "second.yml")
	_ = second

	defaults := &singleFileConfig{}

	mgr := Initialize("empty", defaults)
	a.So(mgr, should.NotBeNil)

	mgr.Parse("--config", first)
	err := mgr.ReadInConfig()
	a.So(err, should.BeNil)

	res := new(singleFileConfig)
	err = mgr.Unmarshal(res)
	a.So(err, should.BeNil)

	a.So(res.Foo, should.Resemble, "10")
	a.So(res.Bar, should.Resemble, map[string]string{
		"a": "baz",
		"b": "quu",
	})
}

type multiFileConfig struct {
	ConfigPaths []string          `name:"config"`
	Foo         string            `name:"foo"`
	Bar         map[string]string `name:"bar"`
}

func TestReadMultiConfig(t *testing.T) {
	a := assertions.New(t)

	_, filename, _, _ := runtime.Caller(0)

	first := path.Join(filepath.Dir(filename), "first.yml")
	second := path.Join(filepath.Dir(filename), "second.yml")
	_ = second

	defaults := &multiFileConfig{}

	mgr := Initialize("empty", defaults)
	a.So(mgr, should.NotBeNil)

	mgr.Parse("--config", first, "--config", second)
	err := mgr.ReadInConfig()
	a.So(err, should.BeNil)

	res := new(multiFileConfig)
	err = mgr.Unmarshal(res)
	a.So(err, should.BeNil)

	a.So(res.Foo, should.Resemble, "20")
	a.So(res.Bar, should.Resemble, map[string]string{
		"a": "baz",
		"b": "hey",
		"c": "yo!",
	})
}
