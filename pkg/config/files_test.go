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

package config

import (
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
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

	mgr := InitializeWithDefaults("empty", "empty", defaults)
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

	mgr := InitializeWithDefaults("empty", "empty", defaults)
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
