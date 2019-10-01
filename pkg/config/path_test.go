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

package config

import (
	"os"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type configWithPath struct {
	ConfigPaths []string `name:"config" shorthand:"c" description:"The location of the config path"`
	DataDir     string   `name:"data-dir" description:"The location of the data dir"`
}

type configWithoutPath struct{}

func TestConfigPathHome(t *testing.T) {
	a := assertions.New(t)

	os.Unsetenv("PWD")
	os.Setenv("HOME", "/home/johndoe")
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("TEST_CONFIG")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$HOME/.test.yml]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/home/johndoe/.test.yml"})
}

func TestConfigPathXDGHome(t *testing.T) {
	a := assertions.New(t)

	os.Unsetenv("PWD")
	os.Unsetenv("HOME")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Unsetenv("TEST_CONFIG")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$XDG_CONFIG_HOME/test/test.yml]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/home/johndoe/.config/test/test.yml"})
}

func TestConfigPathDefault(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("TEST_CONFIG")

	defaults := &configWithPath{
		ConfigPaths: []string{
			"/env/test.yml",
			"/quu/qux.yml",
		},
	}

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[/env/test.yml,/quu/qux.yml]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/env/test.yml", "/quu/qux.yml"})
}

func TestConfigPathEnv(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "/foo/bar/baz.yml")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/foo/bar/baz.yml"})
}

func TestConfigPathCliFlag(t *testing.T) {
	a := assertions.New(t)

	os.Unsetenv("PWD")
	os.Unsetenv("HOME")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Unsetenv("TEST_CONFIG")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$XDG_CONFIG_HOME/test/test.yml]")

	config.Parse("--config", "/foo/bar", "--config", "/quu/qux")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/foo/bar", "/quu/qux"})
}

func TestConfigPathDefine(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("PWD", "/home/johndoe")
	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Unsetenv("TEST_CONFIG")

	defaults := &configWithoutPath{}

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$PWD/.test.yml,$HOME/.test.yml,$XDG_CONFIG_HOME/test/test.yml]")
}

func TestDataDirHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Unsetenv("XDG_DATA_HOME")
	os.Unsetenv("TEST_DATA_DIR")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/home/johndoe/.test")
}

func TestDataDirXDGHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Unsetenv("TEST_DATA_DIR")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/home/johndoe/.data/test")
}

func TestDataDirDefault(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Unsetenv("TEST_DATA_DIR")

	defaults := &configWithPath{
		DataDir: "/var/run/test",
	}

	config := InitializeWithDefaults("test", "test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/var/run/test")
}

func TestDataDirEnv(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Setenv("TEST_DATA_DIR", "/foo/bar")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/foo/bar")
}

func TestDataDirCliFlag(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Unsetenv("TEST_DATA_DIR")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", "test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse("--data-dir", "/mnt/data")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/mnt/data")
}

func TestDataDirDefine(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Unsetenv("TEST_CONFIG")

	defaults := &configWithoutPath{}

	config := InitializeWithDefaults("test", "test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("data-dir")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "$XDG_DATA_HOME/test")
}
